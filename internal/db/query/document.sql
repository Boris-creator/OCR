-- name: GetDocumentByHash :one
SELECT id, ocr FROM documents
WHERE hash = $1 AND chat_id = $2;

-- name: CreateDocument :one
INSERT INTO documents (
    file_id, chat_id, hash, ocr
) VALUES(
    $1, $2, $3, $4
) RETURNING id;