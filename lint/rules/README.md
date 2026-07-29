# Lint Rules

This directory contains lint rules for the header CLI.

## Primary vs Secondary Consumers

When writing lint rules, it is important to distinguish between **primary consumers** and **secondary consumers** of metadata standards.

- **Primary Consumers**: These are the platforms that defined the structure and standard. Examples include Meta (Facebook) for OpenGraph (`og:*`) and Twitter (X) for Twitter Cards (`twitter:*`). Rules named after these platforms are acceptable (e.g., `facebookRule`, `twitterRule`, `opengraphRule`).
- **Secondary Consumers**: These are platforms or tools that utilize the standards defined by primary consumers, but do not define the standard themselves. Examples include Discord, LinkedIn, Slack, etc.

### Guidelines for Secondary Consumers

- **Do not create specific rules for secondary consumers** (e.g., do not create a `discordRule`). Secondary consumers should only be acknowledged in the comments. A primary consumer is Meta / Twitter as they defined the structure/standard, secondary are people that just use the standard.
- If a secondary consumer requires a specific combination of tags to work correctly (for example, Discord requiring `twitter:player` for inline video playback of `og:video` content), this check should be integrated into the relevant primary standard's rule (e.g., `twitterRule` or `opengraphRule`).
- You may acknowledge secondary consumers in the comments of the primary rule to explain *why* a particular check is beneficial.
