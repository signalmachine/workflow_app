# v2 Client And Multimodal Surfaces

Date: 2026-03-21
Status: Draft v2 client-surface note
Purpose: capture the planned mobile, web, and multimodal client breadth that should land after thin-v1 foundations are complete, while separating that v2 breadth from the minimum thin-v1 client-ingress foundation required for real user testing.

## 1. Review conclusion

The current thin-v1 codebase has credible business-domain foundations for later web and mobile clients, but it does not yet have complete client-platform foundations for multimodal usage.

What is already strong enough to carry into later client work:

1. device-scoped sessions and tenant-safe org context
2. AI run, step, artifact, recommendation, and delegation persistence
3. shared document, approval, accounting, inventory, execution, and reporting domain boundaries
4. support-depth party and contact records intended for cross-module reuse

What is not yet implemented in the active codebase:

1. a real HTTP or API application surface
2. attachment persistence plus bounded upload and download contracts
3. a first-class inbound message model for text, transcript, image, audio, and attachment references
4. conversation-oriented ingestion for user requests arriving from web or mobile clients
5. explicit speech-transcript, OCR, or multimodal processing boundaries
6. a durable queued request model linking inbound user input to asynchronous AI processing and later human action

## 2. v2 client objective

V2 may add deliberate client-surface breadth on top of thin-v1 foundations:

1. a Flutter mobile client
2. richer operational web surfaces
3. multimodal request submission through text, voice, image, and attachment combinations
4. local-language user interaction with backend processing remaining language-agnostic where practical

The intended flow is:

1. the client captures text or voice plus optional images or documents
2. the client sends the message and attachment references through explicit backend contracts
3. the backend resolves transcript, OCR, attachment, and AI-routing work through normal domain-service boundaries
4. approved downstream actions still use the same document, approval, posting, inventory, and execution foundations

## 3. Thin-v1 readiness assessment

Thin v1 likely avoids a rewrite of the business core for later web and mobile support, but it does not yet fully support the future multimodal client shape by itself.

That means:

1. thin v1 is strong enough in domain foundations
2. thin v1 is not yet complete in client-ingress and multimodal platform foundations
3. v2 can safely deepen client breadth only if thin v1 first lands the minimum queued request and browser-testing seams needed to validate real user workflows

## 4. Thin-v1 foundation boundary

The minimal client-ingress slice is now part of thin-v1 foundation rather than an optional promotion.

That thin-v1 slice should stay narrow:

1. one stable backend API shape suitable for web first and mobile later
2. session-auth flows usable by browser clients
3. bounded attachment upload and download contracts
4. one minimal persist-first request-ingest path for typed text plus optional attached images or files
5. queued AI processing with durable request status rather than immediate-response as the primary interaction model
6. one minimal web review surface or API path sufficient to test the inbound request -> AI proposal -> document -> approval -> posting or execution loop

That thin-v1 slice should not become:

1. a broad operational web UI
2. full mobile-product work
3. a polished chat product
4. a broad conversation-history system unless required for foundation correctness

## 5. Recommended decision rule

Use this rule for client-surface scope:

1. keep the minimum persist-first web-ingress, attachment, queued-AI, and review path in thin v1 because it is important, foundation-level, and feasible without broad product-surface expansion
2. keep richer multimodal capture, polished conversation UX, full mobile-product depth, and broader operational web surfaces in v2
3. do not let thin-v1 client foundation expand into a broad chat product or a broad operational UI
