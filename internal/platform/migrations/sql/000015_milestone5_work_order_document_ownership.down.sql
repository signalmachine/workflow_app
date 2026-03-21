DROP TABLE IF EXISTS work_order_document_ids_to_remove;

CREATE TEMP TABLE work_order_document_ids_to_remove AS
SELECT document_id
FROM work_orders.documents;

DROP TABLE IF EXISTS work_orders.documents;

DELETE FROM documents.documents d
WHERE d.type_code = 'work_order'
  AND EXISTS (
  	SELECT 1
  	FROM work_order_document_ids_to_remove ids
  	WHERE ids.document_id = d.id
  );
