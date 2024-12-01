# Use the official Rust image as the base image
FROM rust:latest

# Set the working directory inside the container
WORKDIR /usr/src/app

# Copy the Cargo.toml and Cargo.lock files
COPY Cargo.toml Cargo.lock ./

# Copy the source code
COPY src ./src

RUN apt-get update && apt-get install -y cmake


# Build the application
RUN cargo build --release

# Expose the port that the application will run on
EXPOSE 8080

# Run the application
CMD ["./target/release/magpie"]