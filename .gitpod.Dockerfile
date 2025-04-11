FROM gitpod/workspace-full

# Install additional tools
USER root

# Install Docker
RUN curl -fsSL https://get.docker.com | sh

# Add user to docker group
RUN usermod -aG docker gitpod

# Reset user to gitpod
USER gitpod