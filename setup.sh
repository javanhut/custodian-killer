#!/bin/bash

# Custodian Killer Project Setup Script 🦍
echo "🦍 Setting up Custodian Killer project structure..."

# Create directory structure
mkdir -p templates
mkdir -p aws
mkdir -p storage
mkdir -p reports
mkdir -p utils
mkdir -p web/static

echo "📁 Directory structure created!"

# Create go.mod if it doesn't exist
if [ ! -f "go.mod" ]; then
    echo "📦 Initializing Go module..."
    go mod init custodian-killer
fi

# Update go.mod with dependencies
echo "📥 Adding dependencies..."
go mod tidy

echo "🔧 Building project..."
go build -o custodian-killer

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    echo "🚀 Ready to run:"
    echo "   ./custodian-killer          # Interactive mode"
    echo "   ./custodian-killer --help   # See all commands"
    echo "   ./custodian-killer version  # Check version"
    echo ""
    echo "🎯 Try creating your first policy:"
    echo "   ./custodian-killer"
    echo "   > make policy"
    echo ""
else
    echo "❌ Build failed. Check the errors above."
    exit 1
fi
