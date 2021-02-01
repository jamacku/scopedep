build:
	go build

install: build
	install -o root -g root scopedep /usr/bin
	install -o root -g root init/scopedep.service /usr/lib/systemd/system
	systemctl daemon-reload

uninstall:
	-systemctl stop scopedep.service
	rm -f /usr/bin/scopedep
	rm -f /usr/lib/systemd/system/scopedep.service
	systemctl daemon-reload
