# GoWebApp
This is the repository for my web Application project.

- Built in Go 1.22.5
- Uses the [chi router](https://github.com/go-chi/chi)
- Uses [alex edwards SCS](https://github.com/alexedwards/scs/v2) session management
- Uses [nosurf](https://github.com/justinas/nosurf)
- Uses [PostgreSQL](https://www.postgresql.org/download/windows/) 16.4
- Uses [Sodapop](https://gobuffalo.io/documentation/database/pop/) 6.1.1


# How to run the application
Compile and run the application using the ./run.sh script

# Soda
This command creates a new up/down migration with the name you specified

soda generate fizz <migration_name>

This command below takes all migrations upwards

soda migrate

This command below migrates one step downwards (not all at once)

soda migrate down

This command below resets the database to a clean slate with all migrations in place
Note: Make sure no other connections exist for the database.

soda reset

# PostgreSQL
Note to self: Never use the default postgres database for applications.
Create one specifically for your application, like here below "bookings"

psql -h winhost -p 5432 -U postgres -d bookings

DROP DATABASE bookings