# Ledger API Readme

This is an API implementation for managing ledger transactions using Go. It is built on top of the Gorilla web toolkit.

## Requirements
- Go 1.16+
- Gorilla web toolkit v1.8.0
- Shopspring Decimal package v1.6.0

## Installation
1. Clone the repository `git clone https://github.com/tebrizetayi/ledgerservice.git`
2. Install the required packages using `go mod download`

## Usage
1. To start the server, run `docker-compose up -d`
2. The server runs on `http://localhost:8080` by default.
3. Available endpoints:
   - `POST /users/{uid}/add`: Adds a new transaction for the user specified by `uid`
    
    ``` curl -X POST   -H "Content-Type: application/json"   -d '{"amount": 100, "idempotency_key": "123e4567-e89b-12d3-a456-426614174001"}'   http://localhost:8080/users/123e4567-e89b-12d3-a456-426614174000/add ```

   - `GET /users/{uid}/balance`: Retrieves the balance of the user specified by `uid`
   ``` http://localhost:8080/users/123e4567-e89b-12d3-a456-426614174000/balance```
   - `GET /users/{uid}/history`: Retrieves the transaction history of the user specified by `uid`
    ``` http://localhost:8080/users/123e4567-e89b-12d3-a456-426614174000/history```
4. To run the tests, run `go test ./... -v`
5. To stop the server, run `docker-compose down`
6. There are test users with the following IDs:
   - `123e4567-e89b-12d3-a456-426614174000`
   - `123e4567-e89b-12d3-a456-426614174001`
   - `123e4567-e89b-12d3-a456-426614174002`

## API Documentation

### TransactionManager
- `AddTransaction(ctx context.Context, transaction transactionmanager.Transaction) (transactionmanager.Transaction, error)`: Adds a new transaction to the ledger.
- `GetUserBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)`: Retrieves the balance of the specified user.
- `GetUserTransactionHistory(ctx context.Context, userID uuid.UUID, page int, pageSize int) ([]transactionmanager.Transaction, error)`: Retrieves the transaction history of the specified user.

### Controller
- `GetUserBalance(w http.ResponseWriter, r *http.Request)`: Retrieves the balance of a user.
- `AddTransaction(w http.ResponseWriter, r *http.Request)`: Adds a new transaction to the ledger.
- `GetUserTransactionHistory(w http.ResponseWriter, r *http.Request)`: Retrieves the transaction history of a user.

### AddTransactionRequest
- `Amount float64`: The amount of the transaction.
- `IdempotencyKey uuid.UUID`:It guarantees that caller will call exactely once for the same money transfer.


## TransactionRepository.AddTransaction Function Explanation

The `AddTransaction` function in the `TransactionRepository` struct handles adding a transaction while ensuring that the same transaction with same amount is not added multiple times. It does this by using the `Unique(IdempotencyKey,Amount)`. Here is a step-by-step explanation of how the function works:

1. **Begin a new transaction**: A new transaction is started in the database using `t.db.BeginTx(ctx, nil)`. This is important for maintaining consistency and ensuring that multiple operations are executed atomically.

2. **Lock the user row using SELECT FOR UPDATE**: The user row in the `users` table is locked by querying it with the `SELECT ... FOR UPDATE` clause. This prevents other transactions from modifying the user's balance while the current transaction is being processed.

3. **Check for existing user**: If the user is not found, an error `ErrUserNotFound` is returned, and the transaction is rolled back.

4. **Insert the transaction**: The new transaction is inserted into the `transactions` table with its `IdempotencyKey`. If the transaction fails, the database transaction is rolled back and an error is returned.

5. **Update the user's balance**: After successfully inserting the transaction, the user's balance is updated by adding the transaction amount to the current balance. If updating the balance fails, the database transaction is rolled back and an error is returned.

6. **Commit the transaction**: If all the previous steps are successful, the database transaction is committed using `tx.Commit()`. This ensures that all the changes made during this transaction are persisted in the database.

7. **Return the transaction**: Finally, the transaction details are returned, including the `IdempotencyKey`.

    By using the `Unique(IdempotencyKey,Amount)`, it is ensured that the same transaction is not added multiple times. When a request to add a transaction is received, the `Unique(IdempotencyKey,Amount)` is checked to determine if the transaction has already been processed. If it has, the operation is considered idempotent, and the same result is returned without performing the operation again. This helps to maintain consistency in the system and prevents duplicate transactions.

### **Improvement Points:**
- For testing use integresql https://github.com/allaboutapps/integresql
Integresql uses PostgreSQL's pg_create_physical_replication_slot and pg_logical_emit_message to create a new database by cloning an existing one. This approach is faster than traditional methods such as creating a new database from scratch, running migrations, and seeding data.Integresql uses templates in PostgreSQL.

- The use of the migration tool for databases, https://github.com/golang-migrate/migrate, and the process of creating an init container in Kubernetes that will update the database before running the main service.
The golang-migrate/migrate tool is a library and CLI that facilitates database migrations. It supports a variety of databases and offers a simple and efficient way to manage schema changes and data migrations.

### **How would be be deployment process using Migrations tool and K8s?**

Let's discuss the process of using Kubernetes to create an init container that updates the database before running the main service:
- Create a Docker image for your database migration tool: To use golang-migrate/migrate in Kubernetes, you need to create a custom Docker image containing the CLI and your migration files. This image will be used as the init container for your Kubernetes deployment.
- Define the init container in your Kubernetes deployment. This init container should use the custom Docker image. The init container will execute the database migration before the main application starts.
- Configure the main service container: This container will only start once the init container has successfully completed the migration.

By combining the golang-migrate/migrate tool with Kubernetes init containers, we can ensure that your database is updated before your main service starts. This approach provides a robust and automated way to manage database migrations in a containerized environment.

### **Security Concerns:**
To ensure security for the API, we can use JSON Web Tokens (JWT) for authentication and authorization. Here is an overview of how to use JWT with the API:
Implement a login endpoint (or use external Auth0,Okta)where users can authenticate with their credentials (e.g., username and password). Once authenticated, the server generates a JWT token and returns it to the client.
For each subsequent API request, the client must include the JWT token in the Authorization header. The server checks the validity of the token and grants or denies access based on the token's content.
To prevent unauthorized access, we should use HTTPS to encrypt all communication between the client and server.


## Answer to questions:

## **How could the service be developed to handle thousands of concurrent users with hundreds of transactions each?**

To improve the performance of `AddTransaction`, `GetUserBalance`, and `GetUserTransactionHistory` functions, consider implementing the following strategies:

### AddTransaction:

- Use a distributed lock to allow only one instance to execute the function.
- Employ batch inserts for multiple transactions.
- Utilize connection pooling for database connections.
- Decouple transaction insertion from the HTTP request/response using a message queue.

### GetUserBalance:

- Cache the balance for a set period to reduce database queries.
- Apply connection pooling for efficient database connection management.

### GetUserTransactionHistory:

- Implement pagination to limit returned transactions per query.
- Cache transaction history for a certain duration to minimize database queries.
- Use connection pooling for efficient database connection management.


## **What has to be paid attention to if we assume that the service is going to run in multiple instances for high availability?**
## High Availability Considerations

In my opinion, to achieve high availability for the service it should be deployed to cloud(GCP,AWS,Azure) and for example in the Google Cloud Platform (GCP), one could consider implementing the following strategies and leveraging various GCP services:

Redundancy: I would suggest deploying multiple instances of the addtransaction service across different zones or regions using Google Kubernetes Engine (GKE) or Cloud Run. This approach would provide redundancy and help maintain service availability in case of instance or zone failures.

Load balancing: In my view, employing GCP's Cloud Load Balancing to distribute incoming transaction requests among the available instances of the addtransaction service would be beneficial. This would help prevent overloading a single instance and improve overall performance and reliability.

Data replication: I would recommend using Google Cloud Spanner or Cloud SQL for the addtransaction service database, as they offer built-in data replication and high availability across multiple zones or regions. This would help ensure data consistency and accessibility.

Fault tolerance: To handle failures gracefully, I believe designing the addtransaction service with features such as retries, timeouts, and circuit breakers would be valuable. Asynchronous processing could also be implemented using Cloud Functions or Cloud Pub/Sub to improve fault tolerance.

Monitoring and health checks: I would advise utilizing Google Cloud Monitoring and Logging to monitor the performance, health, and error rates of the addtransaction service and its underlying infrastructure. Setting up health checks with Cloud Load Balancing or GKE would also be useful for detecting and reporting potential issues.

Scalability: In my opinion, designing the addtransaction service for scalability by leveraging GKE's autoscaling, Cloud Run's automatic scaling, or App Engine's scaling features would be beneficial. This would ensure that the service can handle increased demand by adjusting the number of instances automatically.

By adopting these strategies and making use of GCP services, I believe high availability for the service can be achieved, ensuring uninterrupted access and operation for users even in the face of failures or increased demand.


### **How does the the add endpoint have to be designed, if the caller cannot guarantee that it will call exactely once for the same money transfer?**
The idempency key is used to ensure that the same transaction is not added multiple times. When a request to add a transaction is received, the `Unique(IdempotencyKey,Amount)` is checked to determine if the transaction has already been processed. If it has, the operation is considered idempotent, and the same result is returned without performing the operation again. This helps to maintain consistency in the system and prevents duplicate transactions.






