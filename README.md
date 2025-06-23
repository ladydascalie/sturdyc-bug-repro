# sturdyc-bug-repro

Should be very simple to use:

```
docker compose up -d

go run .
```

## with: github.com/viccon/sturdyc v1.1.4 (WHAT I DID NOT WANT)

last get or fetch returns `err: sturdyc: invalid response type`

## with: github.com/viccon/sturdyc v1.1.5 (WHAT I DO WANT)

last get or fetch returns `err: leaderboard not found`
