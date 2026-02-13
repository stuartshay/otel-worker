FROM mcr.microsoft.com/devcontainers/go:2-1.24-bookworm

# Remove Yarn repository with expired GPG key before features install
# This prevents apt-get update failures during feature installation
RUN rm -f /etc/apt/sources.list.d/yarn.list
