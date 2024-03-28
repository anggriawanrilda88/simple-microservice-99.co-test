# Backend Tech Challenge
An exercise to create backend development and microservices architecture.

This Microservices use Golang to demonstrate a grasp of the principles of microservice architecture. 

There are 3 Microservices created on this test:
- **Listing service:** (Default Code As Example To Create Other Microservices, Use Python Programming Language) Stores all the information about properties that are available to rent and buy.
- **User service:** (The Test Microservice) Stores information about all the users in the system, See quick setup to know how to run this service.
- **Public API layer:** (The Test Microservice) Set of APIs that are exposed to the web/public, See quick setup to know how to run this service.


# Quick Setup
To run service user API and service public API follow this step:
```bash
# Run the user service
cd user_service

# Set port on main.go file
go run main.go 
```

```bash
# Run the public API service
cd public_api_service

# Set port on main.go file
go run main.go
```