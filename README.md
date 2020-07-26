# keycloac-oidc-connect-app
EXAMPLE: Go application connected OIDC authentification using KeyCloak

## Caution

I implemented this repository by follow articles.

https://qiita.com/shibukawa/items/fd78d1ca6c23ce2fa8df

## Usage

docker run -d -p 18080:8080 -e KEYCLOAK_USER=admin -e KEYCLOAK_PASSWORD=admin --name keycloak jboss/keycloak
go run main.go

[Browser]
http://localhost:8080

input: admin/admin

Then you can get login success.
