# v2 Client And Multimodal Surfaces

Date: 2026-03-21
Status: Draft v2 client-surface note
Purpose: capture the planned mobile and multimodal client breadth that should land after the promoted v1 web and AI foundations are complete, while separating that later breadth from the active v1 web-layer and backend work.

## 1. Review conclusion

The current thin-v1 codebase has credible business-domain foundations for later web and mobile clients, but it does not yet have complete client-platform foundations for multimodal usage.

What is already strong enough to carry into later client work:

1. device-scoped sessions and tenant-safe org context
2. AI run, step, artifact, recommendation, and delegation persistence
3. shared document, approval, accounting, inventory, execution, and reporting domain boundaries
4. support-depth party and contact records intended for cross-module reuse

What is not yet implemented in the active codebase:

1. a real HTTP or API application surface
2. bounded attachment upload and download contracts on top of the existing attachment persistence model
3. conversation-oriented ingestion for user requests arriving through an actual API surface
4. explicit speech-transcript, OCR, or multimodal processing boundaries beyond the current transcription-derived record baseline

Promotion note:

1. the minimum backend API surface, browser-usable session-auth path, and bounded attachment upload and download contract are now promoted into active v1 planning because a usable AI-agent-first v1 should not stop at service-layer seams only
2. the usable web application layer itself is now also promoted into active v1 planning
3. richer multimodal capture, polished conversation UX, and full mobile depth remain v2

## 2. v2 client objective

V2 may add deliberate client-surface breadth on top of the promoted v1 web and AI foundations:

1. a Flutter mobile client
2. multimodal request submission through text, voice, image, and attachment combinations
3. local-language user interaction with backend processing remaining language-agnostic where practical

The intended flow is:

1. the client captures text or voice plus optional images or documents
2. the client sends the message and attachment references through explicit backend contracts
3. the backend resolves transcript, OCR, attachment, and AI-routing work through normal domain-service boundaries
4. approved downstream actions still use the same document, approval, posting, inventory, and execution foundations

Backend rule:

1. the future mobile client and the promoted web layer should share one backend foundation rather than diverging into separate product backends
2. differences between web and mobile should mostly live in client interaction design, capture flow, and presentation rather than in duplicated domain logic or duplicate backend truth models

## 3. Thin-v1 readiness assessment

Thin v1 likely avoids a rewrite of the business core for later web and mobile support, but it does not yet fully support the future multimodal client shape by itself.

That means:

1. thin v1 is strong enough in domain foundations
2. v1 is not yet complete in web, API, and multimodal platform foundations
3. v2 mobile and multimodal breadth should deepen only after v1 lands the promoted usable web layer and shared backend contracts

## 4. Thin-v1 foundation boundary

The client-ingress slice and usable web layer are now part of active v1 foundation rather than optional promotions.

That promoted v1 slice should cover:

1. one stable backend API shape suitable for web first and mobile later
2. session-auth flows usable by browser clients
3. bounded attachment upload and download contracts
4. one persist-first request-ingest path for typed text plus optional attached images or files
5. queued AI processing with durable request status rather than immediate-response as the primary interaction model
6. a usable web application layer for review, approval, and downstream inspection

Promoted conclusion:

1. these backend, attachment-transport, and usable-web-layer foundations should now be treated as active v1 work rather than left only in this v2 note

That promoted v1 slice should not become:

1. full mobile-product work
2. a polished chat product
3. a broad conversation-history system unless required for foundation correctness
4. duplicate backend stacks for web and mobile

## 5. Recommended decision rule

Use this rule for client-surface scope:

1. keep the shared backend contracts plus usable web layer in v1 because the application should be genuinely operable before mobile-product breadth begins
2. keep richer multimodal capture, polished conversation UX, and full mobile-product depth in v2
3. do not let the promoted v1 web layer expand into duplicate web-versus-mobile backends or a broad chat product
