ifeq ($(XDG_CONFIG_HOME),)
	XDG_CONFIG_HOME=$(HOME)/.config
endif
dist/.local/bin/swayext: swayext.go
	go build -o dist/.local/bin/swayext swayext.go
build: dist/.local/bin/swayext
install:
	mkdir -p $(XDG_CONFIG_HOME)/waybar
	stow -S dist -t $(HOME)
uninstall:
	stow -D dist -t $(HOME)
	rmdir $(XDG_CONFIG_HOME)/waybar || :
