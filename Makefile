dist/.local/bin/swayctl: swayctl/*.go
	cd swayctl && go build -o ../dist/.local/bin/swayctl
build: dist/.local/bin/swayctl
install:
	mkdir -p $(HOME)/.config/waybar
	stow -S dist -t $(HOME)
uninstall:
	stow -D dist -t $(HOME)
