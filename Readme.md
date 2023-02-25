# Image service

## Description
Application that accepts, resizes and uploads images to AWS S3

## Requirements:
 * go 1.20
 * docker & docker-compose

## Installation

Clone this repository.

```bash
git clone https://github.com/demius1992/image-service
```

Use docker-compose to run the app

```bash
cd image-service
docker-compose up -d
```

## API:
### GET /health
Used to check if app is running or not

Example Response (Status 200 OK):

```
{"ok"}
```

### GET /health
Used to check if app is running or not

Example Response (Status 200 OK):

```
{"ok"}
