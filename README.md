# GoWebApp
This is the repository for my web Application project. It was something I developed as part of 
following a project course on Udemy for Golang Web Applications

- Built in Go 1.22.5
- Uses the [chi router](https://github.com/go-chi/chi)
- Uses [alex edwards SCS](https://github.com/alexedwards/scs/v2) session management
- Uses [nosurf](https://github.com/justinas/nosurf)
- Uses [PostgreSQL](https://www.postgresql.org/download/windows/) 16.4
- Uses [Sodapop](https://gobuffalo.io/documentation/database/pop/) 6.1.1


# How to run the application
Compile and run the application using the ./run.sh script

# Soda Usage
This command creates a new up/down migration with the name you specified

soda generate fizz <migration_name>

This command below takes all migrations upwards

soda migrate

This command below migrates one step downwards (not all at once)

soda migrate down

This command below resets the database to a clean slate with all migrations in place
Note: Make sure no other connections exist for the database. So exit Dbeaver beforehand etc.

soda reset

# PostgreSQL

This application was developed with a PostgreSQL database as part of its backend.

Note to self: Never use the default postgres database for applications.
Create one specifically for the application you are developing, like here below "bookings"

Example usage:
psql -h winhost -p 5432 -U postgres -d bookings

DROP DATABASE bookings