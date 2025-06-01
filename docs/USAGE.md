I'll help you create a Markdown file with the proper content. To create a new file at `docs/USAGE.md`, follow these steps:

1. First make sure the `docs` directory exists
2. Create a new file in that directory called `USAGE.md`
3. Add the following content to it:

```markdown
# Usage Guide

## 1. Start Services

Start Redis, API, and Worker using Docker Compose:

```sh
docker compose up -d
```

## 2. Check Running Containers

List running containers and get their names:

```sh
docker compose ps
```

## 3. Test Redis

Check Redis connectivity from inside the container (replace `<redis-container-name>` with the actual name from the previous step):

```sh
docker exec -it <redis-container-name> redis-cli ping
# Should return: PONG
```

## 4. Test API

Submit a job to the API (adjust endpoint and payload as needed):

```sh
curl -X POST http://localhost:8080/jobs -d '{"task":"example"}' -H "Content-Type: application/json"
```

## 5. Check Worker Logs

Verify that the worker is processing jobs (replace `<worker-container-name>`):

```sh
docker logs <worker-container-name>
```

## 6. (Optional) Print All Jobs in Queue

If you have a debug endpoint or can run `PrintAllJobs`, use it to list jobs in the queue.
```

You can create this file using the GoLand IDE by right-clicking on the `docs` directory, selecting "New" > "File", entering `USAGE.md`, and then pasting the content.