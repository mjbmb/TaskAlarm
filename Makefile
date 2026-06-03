BINARY     := taskalarm
INSTALL_DIR := $(HOME)/.local/bin
AUTOSTART_DIR := $(HOME)/.config/autostart
BINARY_PATH := $(INSTALL_DIR)/$(BINARY)

.PHONY: build test install uninstall

build:
	go build -o $(BINARY) .

test:
	go test ./features/...

install: build
	@mkdir -p $(INSTALL_DIR) $(AUTOSTART_DIR)
	cp $(BINARY) $(BINARY_PATH)
	@printf '[Desktop Entry]\nType=Application\nName=TaskAlarm\nComment=Tagesplaner mit Erinnerungen\nExec=$(BINARY_PATH)\nIcon=appointment-soon\nX-GNOME-Autostart-enabled=true\nX-GNOME-Autostart-Delay=5\nStartupNotify=false\n' \
		> $(AUTOSTART_DIR)/$(BINARY).desktop
	@echo "✓ Installiert: $(BINARY_PATH)"
	@echo "✓ Autostart:   $(AUTOSTART_DIR)/$(BINARY).desktop"

uninstall:
	rm -f $(BINARY_PATH) $(AUTOSTART_DIR)/$(BINARY).desktop
	@echo "✓ Deinstalliert"
