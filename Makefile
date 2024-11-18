swayctl/bin/swayctl: swayctl/*.go
	cd swayctl && go build -o bin/swayctl
build: swayctl/bin/swayctl
install:
	mkdir -p $(HOME)/.config/waybar
	stow -S dist -t $(HOME)
	sudo cp swayctl/bin/swayctl /usr/bin
uninstall:
	stow -D dist -t $(HOME)
	sudo rm /usr/bin/swayctl
