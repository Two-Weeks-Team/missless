# Dev.to Upload Workflow

This guide explains how to upload your daily devlogs to dev.to for the Gemini Live Agent Challenge.

## Requirements

Every post MUST include:
- The sentence: "I created this post for the purposes of entering the Gemini Live Agent Challenge."
- The hashtag: `#GeminiLiveAgentChallenge`

## Method 1: Manual Web Upload

1. Go to [dev.to/new](https://dev.to/new).
2. Copy the content from your generated devlog file (e.g., `docs/devlog/2026-02-25.md`).
3. Paste it into the editor.
4. Add relevant tags (e.g., `GeminiLiveAgentChallenge`, `DevLog`, `BuildInPublic`).
5. Click **Save as Draft** or **Publish**.

## Method 2: CLI/API Upload (Recommended)

You can use the local skill script to upload your devlog directly from the terminal. This requires a `DEVTO_API_KEY`.

### Setup

Ensure you have your `DEVTO_API_KEY` exported in your environment:
```bash
export DEVTO_API_KEY=your_api_key_here
```

### Usage

Run the following command to upload a devlog file:

```bash
python3 ~/.claude/skills/devto-daily-devlog-uploader/scripts/upload_devlog.py --file <path_to_devlog> [--title "Your Title"] [--tags "tag1,tag2"]
```

## Asking Claude to Help

You can ask Claude to handle the creation and upload process for you.

### Trigger Phrases

- "Create a daily devlog for today and upload it to dev.to"
- "Generate a devlog from my last 24h activity and post it as a draft on dev.to"
- "Upload docs/devlog/2026-02-25.md to dev.to"

---
I created this post for the purposes of entering the Gemini Live Agent Challenge.
#GeminiLiveAgentChallenge
