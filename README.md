# mastostart
Mastostart provides a simple way to login to Mastodon using OAuth2. It is designed to be used as a backend for a web app.

## Dev Branch 🎉
This is the dev branch.

## Set up
- REQUIRED: Set up AWS CLI.
- REQUIRED: Run `make deploy` (read the Makefile to see what it does). Make note of the output value for `ApiGateway`. See `redirect_uri` below.
- REQUIRED: Run 'mastostart config jwt-key` to generate an RSA 2048 keypair for JWT signing.
- REQUIRED: Run `mastostart config set --key app_name --value ${app_nam_value}`. Set value to whatever you want call this app.
- REQUIRED: Run `mastostart config set --key website --value ${website_value}`. Set value to the URL of the home/about page for this app.
- REQUIRED: Run `mastostart config set --key redirect_uri --value ${redirect_uri_value}`. Value should be `${ApiGateway}/auth/callback`.
- REQUIRED: Run `mastostart config set --key scopes --value ${csv_of_scopes}`. Value should be a comma-separated list of scopes you want to request from the user. Example: `read,write,follow`.
- OPTIONAL: Run `mastostart config set --key permit_instances --value ${csv_of_instances}`. Value should be a comma-separated list of Mastodon instances (hostnames only) you want to allow users to login to. Leave blank to permit all. Example: `mastodon.social,pleroma.site`.

## Auth Endpoints
- `GET /` - Hello!.
- `GET /auth/callback` - Callback for OAuth2. Retuns a JWT.
  - `?code=${code}` - Required. The OAuth2 code.
  - `?instance_url=${instance_url}` - Required. The Mastodon instance to login to.
- `GET /auth/login` - Setups the app for OAuth2. Redirects to the OAuth2 provider.
  - `?username=${username}` - Required. The username of the user to login as.
  - `?instance_url=${instance_url}` - Required. The Mastodon instance to login to.
- `GET /auth/verify` - Verifies a JWT. Returns the user's Mastodon profile and last status/post.
  - Authorization: Bearer ${jwt}

## General API Endpoints
### Lists
- `GET /api/lists` - Returns a list of the user's lists.
- `GET /api/lists/:listID` - Returns a list.
  - OPTIONAL: `?save=true` - Save the list in the Mastostart database.
  - OPTIONAL: `?public=true` - If saved, make Mastostart-saved list public.

## Instance API Endpoints
- `GET /api/instance` - Returns the instance's info.

