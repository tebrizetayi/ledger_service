API Endpoints
Add Transaction
Add a new transaction for a user.

URL : /users/{uid}/add

Method : POST

Auth Required : NO

Data constraints

json
Copy code
{
    "amount": float,
    "idempotency_key": uuid
}
amount float number, required. Amount of the transaction.
idempotency_key uuid string, required. A unique key for the transaction.
Success Response

Code : 201 CREATED

Content

json
Copy code
{
    "message": "Transaction successfully added"
}
Error Response

400 Bad Request: Invalid user ID or transaction payload
500 Internal Server Error: An error occurred while adding the transaction
Get User Balance
Retrieve the balance of a user.

URL : /users/{uid}/balance

Method : GET

Auth Required : NO

Success Response

Code : 200 OK

Content

json
Copy code
{
    "balance": decimal
}
balance decimal number. The current balance of the user.
Error Response

400 Bad Request: Invalid user ID
500 Internal Server Error: An error occurred while retrieving the user balance
Get User Transaction History
Retrieve the transaction history of a user.

URL : /users/{uid}/history

Method : GET

Auth Required : NO

Query Params

page int, optional. The page number of the results. Defaults to 1.
pageSize int, optional. The number of transactions per page. Defaults to 10.
Success Response

Code : 200 OK

Content

json
Copy code
[
    {
        "id": uuid,
        "user_id": uuid,
        "amount": decimal,
        "created_at": datetime,
        "idempotency_key": uuid
    },
    ...
]
id uuid string. The ID of the transaction.
user_id uuid string. The ID of the user who initiated the transaction.
amount decimal number. The amount of the transaction.
created_at datetime string. The time the transaction was created.
idempotency_key uuid string. The idempotency key for the transaction.
Error Response

400 Bad Request: Invalid user ID or query params
500 Internal Server Error: An error occurred while retrieving the user transaction history
Check User Validity
Check if a user is valid.

URL : /users/{uid}/valid

Method : GET

Auth Required : NO

Success Response

Code : 200 OK

Content

json
Copy code
{
    "valid": boolean
}
valid boolean. true if the user is valid, false otherwise.
Error Response

400 Bad Request: Invalid user ID
500 Internal Server Error: An error occurred while checking the user validity
Docker Compose
To run the app using Docker Compose, run the following command:

css
Copy code
docker-compose up --build
The app will be available at http://localhost:8080/.