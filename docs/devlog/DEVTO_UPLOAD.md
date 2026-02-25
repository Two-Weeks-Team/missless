# Dev.to Upload Workflow

This document describes how to upload your daily devlogs to dev.to using the `devto-daily-devlog-uploader` skill.

## Prerequisites

- **DEVTO_API_KEY**: You must have a dev.to API key set in your environment variables.
- **Claude Skill**: Ensure the `devto-daily-devlog-uploader` skill is installed in your Claude environment.

## Workflow

1. **Write your devlog**: Use the `DAILY_DEVLOG_TEMPLATE.md` as a starting point.
2. **Upload to dev.to**: Ask Claude to upload your devlog.
   - Example: "Upload my devlog 'Day 1: Project Setup' to dev.to as a draft with tags 'gemini,ai,go'."
   - The skill will automatically append the required hackathon sentence and hashtag:
     - `I created this post for the purposes of entering the Gemini Live Agent Challenge.`
     - `#GeminiLiveAgentChallenge`

## Configuration

The skill defaults to `published=false` (Draft). You can explicitly ask to publish it if needed.
