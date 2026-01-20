# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Milestone 1: Module Foundation

**Added**
- Generic service module (`viamdemo:kettle-cycle-test:controller`) with DoCommand stub
- Unit tests for controller lifecycle (NewController, DoCommand, Close)
- Hot reload deployment workflow via `viam module reload-local`
- Makefile with build, test, and packaging targets
- Module metadata (meta.json) for Viam registry integration

**Changed**
- README updated with module structure, Milestone 1 summary, and development instructions

## [0.0.1] - 2026-01-19

### Added
- Project planning documents (product_spec.md, CLAUDE.md)
- Technical decisions for Viam components, data schema, motion constraints
- README target outline with lesson structure
- Claude Code agents for docs, changelog, test review, and retrospectives
