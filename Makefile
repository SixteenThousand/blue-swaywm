ifeq ($(XDG_CONFIG_DIR),)
	CONF_DIR=$(HOME)/.config
else
	CONF_DIR=$(XDG_CONFIG_DIR)
endif
install:
	stow -S . -t $(CONF_DIR)
uninstall:
	stow -D . -t $(CONF_DIR)
