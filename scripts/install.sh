#!/bin/bash
set -e

# Configuration
APP_NAME="telegpt"
INSTALL_DIR="/opt/${APP_NAME}"
BINARY_PATH="${INSTALL_DIR}/${APP_NAME}"
CONFIG_PATH="${INSTALL_DIR}/config.yaml"
SERVICE_FILE="/etc/systemd/system/${APP_NAME}.service"
LOG_DIR="${INSTALL_DIR}/logs"
USER="${APP_NAME}"

# Check if script is run as root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit 1
fi

# Create user if it doesn't exist
if ! id "${USER}" &>/dev/null; then
  echo "Creating ${USER} user..."
  useradd -r -s /bin/false "${USER}"
fi

# Create installation directory
echo "Creating installation directory..."
mkdir -p "${INSTALL_DIR}"
mkdir -p "${LOG_DIR}"

# Copy binary and configuration
echo "Copying files..."
cp "${APP_NAME}" "${BINARY_PATH}"
cp "config.yaml.example" "${CONFIG_PATH}"

# Set permissions
echo "Setting permissions..."
chown -R "${USER}:${USER}" "${INSTALL_DIR}"
chmod 755 "${BINARY_PATH}"
chmod 640 "${CONFIG_PATH}"

# Copy systemd service file
echo "Installing systemd service..."
cp "scripts/${APP_NAME}.service" "${SERVICE_FILE}"
systemctl daemon-reload

echo ""
echo "Installation complete!"
echo ""
echo "Next steps:"
echo "1. Edit ${CONFIG_PATH} with your API keys and configuration"
echo "2. Start the service: systemctl start ${APP_NAME}"
echo "3. Enable on boot: systemctl enable ${APP_NAME}"
echo ""
echo "To check status: systemctl status ${APP_NAME}"
echo "To view logs: journalctl -u ${APP_NAME}"
