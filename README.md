# Go Microservices Project
This project is a basic microservices architecture using Go, Docker, and Docker Compose. It consists of two main services:

- Frontend Service: A simple web server that serves HTML content and Talk to the broker.
- Broker Service: A service that dispatches API requests and returns JSON responses.
- Authentication Service: A service It interacts with a broker to verify user credentials and provides an appropriate response.
  - Workflow
  - User Request: A user sends an authentication request to the broker.
  - Broker Forwarding: The broker forwards the request to the authentication microservice.
  - Authentication Verification: The microservice validates the user's credentials against the stored database.
  - Response: The microservice sends a response back to the broker, indicating whether the authentication was successful or failed.