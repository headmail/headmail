Headmail Helm Chart

This chart deploys:
- headmail backend (binary) exposing public web (/) and admin API (/api)
- optional frontend static server that serves admin web at '/'

Usage examples:
- Install with default values:
  helm upgrade --install headmail chart/headmail

- Override backend image tag:
  helm upgrade --install headmail chart/headmail --set backend.image.tag=0.0.1

- To build and use frontend image from local frontend source:
  1. Build frontend assets and create a docker image that places built files under /app/dist or /usr/share/headmail-frontend.
  2. Set frontend.image.repository and frontend.image.tag to point to that image.
  3. The chart uses an initContainer to copy static files from the frontend image into an nginx container which serves them.
