# Contributing to Storage Bridge

First off, thank you for considering contributing to Storage Bridge! It's people like you that make Storage Bridge such a great tool.

## Where do I go from here?

If you've noticed a bug or have a feature request, make sure to check our [Issues](https://github.com/chriz-3656/storage-bridge/issues) first to see if someone else has already created a ticket. If not, go ahead and [make one](https://github.com/chriz-3656/storage-bridge/issues/new)!

## Fork & create a branch

If this is something you think you can fix, then fork Storage Bridge and create a branch with a descriptive name.

## Get the test suite running

Make sure you have Go 1.22+ installed. 
```bash
go test ./...
```

## Implement your fix or feature

At this point, you're ready to make your changes! Feel free to ask for help; everyone is a beginner at first.

## Make a Pull Request

At this point, you should switch back to your master branch and make sure it's up to date with Storage Bridge's master branch:

```bash
git remote add upstream https://github.com/chriz-3656/storage-bridge.git
git checkout main
git pull upstream main
```

Then update your feature branch from your local copy of main, and push it!

```bash
git checkout -b <your-feature-branch>
git push origin <your-feature-branch>
```

Finally, go to GitHub and make a Pull Request.

## Code Style

- Format your code using `gofmt`
- Run `go vet ./...` to catch common errors
- Make sure all tests pass before submitting

Thank you for your contributions!
