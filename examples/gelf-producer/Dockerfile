FROM ghcr.io/astral-sh/uv:python3.13-alpine3.23

# Set the working directory
WORKDIR /app

# # Install uv
# RUN pip install uv

# Copy the project files
COPY pyproject.toml uv.lock ./

# Install dependencies
RUN uv sync

# Copy the rest of the application
COPY . .

# Set environment variables for GELF
ENV GELF_LOG_LEVEL="INFO"
ENV GELF_SLEEP_PERIOD_MS="500"
ENV GELF_TARGET_HOST="127.0.0.1"
ENV GELF_TARGET_PORT="12201"

# Run the application
CMD ["uv", "run", "."]
