# 0001. Go operator with a Python agent service

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

promoloupe has two very different halves. One reconciles Kubernetes resources,
watches Git and evaluates gates. The other runs an AI agent that calls tools,
reasons over evidence and produces a structured verdict.

The Kubernetes ecosystem for controllers is built around Go: controller-runtime,
Kubebuilder, envtest and client-go. The agent ecosystem is strongest in Python:
the Claude Agent SDK, MCP servers, Pydantic and most evaluation tooling.
Projects such as kagent already combine a Go controller with a Python agent
engine.

## Decision

The controller and the CLI are written in Go using Kubebuilder and
controller-runtime. The reviewer agent is a separate Python service built on
the Claude Agent SDK, managed with uv and checked with ruff, mypy in strict
mode and pytest.

The two halves communicate only through Kubernetes resources (see ADR 0008),
never through direct calls or shared libraries.

## Consequences

Each half uses the idiomatic tools of its ecosystem, which keeps both
approachable for contributors from that ecosystem.

The project carries two toolchains, two CI pipelines and two container images.
Shared types must be defined once (in the CRD schema) and generated or
validated on the Python side, so they cannot drift silently.
