# Panda Game Backend

## Technical Objectives

### Be one with the Zen of Go

Primarily, this means to use Go the Go way.

## Technical Overview

### Auth system

Panda Game the following:

* basic auth to login
* pocketbase for user records, authn and authz
* stores the pocketbase token in a cookie
* allows the Authorization header
