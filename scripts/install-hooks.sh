#!/bin/bash

echo "Installing pre-commit hooks..."

if command -v pre-commit &> /dev/null; then
    pre-commit install
    pre-commit install --hook-type pre-push
    echo "✅ Pre-commit hooks installed successfully!"
else
    echo "❌ pre-commit is not installed. Install it with:"
    echo "   pip install pre-commit"
    echo "   or"
    echo "   brew install pre-commit"
fi
