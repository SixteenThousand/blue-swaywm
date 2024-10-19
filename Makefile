ifeq ($(XDG_CONFIG_DIR),)
	CONF_DIR=$(HOME)/.config/sway
else
	CONF_DIR=$(XDG_CONFIG_DIR)/sway
endif
install:
	stow -S . -t $(CONF_DIR)
uninstall:
	stow -D . -t $(CONF_DIR)
