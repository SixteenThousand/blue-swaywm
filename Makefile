IGNORE="--ignore=Makefile"
ifeq ($(XDG_CONFIG_DIR),)
	CONF_DIR=$(HOME)/.config
else
	CONF_DIR=$(XDG_CONFIG_DIR)
endif
install:
	mkdir -p $(CONF_DIR)/waybar
	stow -S . -t $(CONF_DIR) $(IGNORE)
uninstall:
	stow -D . -t $(CONF_DIR) $(IGNORE)
