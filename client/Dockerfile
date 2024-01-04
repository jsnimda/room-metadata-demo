# Use an official Node runtime as a parent image
FROM node:18.18.2-alpine

# Set the working directory inside the container
WORKDIR /app

RUN npm i -g @ionic/cli

# Copy the package.json and package-lock.json files
COPY package*.json ./

# Install all dependencies, including development ones
RUN npm install

# Copy the rest of the client's code
COPY . .

# Make port 8100 available to the world outside this container
EXPOSE 8100

# Define environment variable
ENV IONIC_SERVE_HOST=0.0.0.0

# Run ionic serve when the container launches
CMD ["ionic", "serve", "--no-open", "--port", "8100", "--host", "0.0.0.0"]
