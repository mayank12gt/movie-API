## Introduction
This is a REST API backend for a movie information service with a Postgres database with pagination, filtering, stateful user-auth, email verification and permission based access control, deployed on AWS EC2 behind a Caddy reverse proxy.
## Permissions
The service has 2 level of permissions "movies:read" and "movies:write". While signing up "movies:read" permission is granted to users, such users cannot perform request that require "movies:write" permission. "movies:write" permission can be granted only by the owner by directly accessing the database, this is to provide a level of security to the database
## User Activation
After signing up a verification email with a validation token is sent to the respective email account. To activate the account, another POST request to /users/activate must be sent with this token. this is to prevent users from submitting an invalid/inactive email address.
## User Authentication
Some request like GET all movies, GET average movie rating are public and do not require authentication while others like GET/POST user rating for a movie require user authentication.
The api uses a stateful auth system, where to authenticate users have to send the email and passsword to the /users/authenticate endpoint, if the email exists and password is correct, an auth token with 24 hour validity is generated to track the user's session and sent with the response as a URL cookie. For requests which require authentication, this token must be sent with the request





## ScreenShots:
<img width="960" alt="movie-api1" src="https://github.com/mayank12gt/movie-web-app/assets/96809211/1b373898-b026-489c-9aa8-e2c56a12e4dc">
<img width="960" alt="movie-api2" src="https://github.com/mayank12gt/movie-web-app/assets/96809211/c29fdd8d-148e-478b-92fb-715f99d58f1a">
<img width="960" alt="m" src="https://github.com/mayank12gt/movie-web-app/assets/96809211/f78e9010-48b5-4575-ac72-62e1887a9f11">
