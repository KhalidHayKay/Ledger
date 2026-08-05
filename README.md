# Ledger

A Payment Gateway service between a fictional e-commerce platform (FicMart) and a fictional Bank.

## Overview

Ledger is a work-in-progress Go service that sits between FicMart and a bank integration layer. It is designed to accept payment intents, communicate with the bank, manage payment state, and handle retries and terminal failures.

## Repository Scope

This repository currently includes:

- core payment intent domain logic
- bank integration abstractions
- payment event tracking
- idempotency support
- worker and queue components
- configuration, database, and notifier plumbing

## Status

This project is not complete yet. The repository is a minimal starting point for development and documentation will expand as the implementation matures.

## Getting Started

The repository contains Docker and Compose files, plus Go entry points under `cmd/`.

If you want to explore the code, the main areas are:

- `cmd/` – application, CLI, and worker entry points
- `internal/domain/` – payment gateway domain logic
- `internal/jobs/` – background processing and queue workers
- `internal/platform/` – configuration, database, middleware, notifier, and rendering support
