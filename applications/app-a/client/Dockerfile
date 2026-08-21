FROM node:22-alpine
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
EXPOSE 4201
CMD ["npm", "start", "--", "--host", "0.0.0.0", "--port", "4201"]
