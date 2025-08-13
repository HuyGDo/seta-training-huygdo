# ---- Base Stage ----
# Use a specific Node.js version for reproducibility
FROM node:18-alpine AS base

# Set the working directory in the container
WORKDIR /usr/src/app

# Copy package.json and package-lock.json to leverage Docker cache
COPY package*.json ./

# ---- Dependencies Stage ----
# Install all dependencies, including devDependencies for potential build steps
FROM base AS dependencies
RUN npm install

# ---- Production Stage ----
# This is the final, lean image
FROM base AS production

# Copy only the production dependencies from the 'dependencies' stage
COPY --from=dependencies /usr/src/app/node_modules ./node_modules

# Copy the rest of the application source code
COPY . .

# Expose the port the app runs on
EXPOSE 4000

# Define the command to run the application in production
CMD ["node", "server.js"]