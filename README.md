# Go Microservices Project
This project is a basic microservices architecture using Go, Docker, and Docker Compose. It consists of two main services:

- **Frontend Service**: A simple web server that serves HTML content and Talk to the broker.
- **Broker Service**: A service that dispatches API requests and returns JSON responses.
- **Authentication Service**: A service It interacts with a broker to verify user credentials and provides an appropriate response.
  - *Workflow*
  - User Request: A user sends an authentication request to the broker.
  - Broker Forwarding: The broker forwards the **request** to the authentication microservice.
  - Authentication Verification: The microservice validates the user's credentials against the stored database.
  - Response: The microservice sends a response back to the broker, indicating whether the authentication was successful or failed.
- **Logger Service**: Log events and information from other microservices within a distributed system. It acts as a centralized logging solution.
  - Centralized Logging: Collects logs from multiple microservices.
  - Database Storage: Stores logs in a MongoDB database for easy retrieval and analysis.
  - Accessibility: Accessible to all microservices within the cluster, but not directly exposed to the internet.
- **Mail Service**: This microservice is responsible for sending emails within the distributed system. It acts as a centralized email gateway.
- **Listner Service with RabbitMQ**: a listener service and RabbitMQ to enable asynchronous, decentralized communication between microservices.
    - Broker Role: The broker acts as a message broker, forwarding requests to a queue managed by RabbitMQ.
    - Listener Service: The listener service monitors the queue, retrieves requests, and directs them to the appropriate microservice.
    - Example: A log request is forwarded to the queue, retrieved by the listener, and processed by the logger service.