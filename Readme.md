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

### POST /images
Used to check if app is running or not

Example Response (Status 200 OK):

```
{
    "createdAt": "2023-03-03T14:19:10Z",
    "id": "4c4ac123-945c-4840-9479-878886da04e3",
    "url": "http://localstack:4566/my-bucket/4c4ac123-945c-4840-9479-878886da04e3?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=test%2F20230303%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20230303T141909Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=9ee561678ce84a89fe9edb9703d6d138eac3d9881dd224629996afe776028f7d"
}
