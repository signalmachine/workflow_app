DROP TRIGGER IF EXISTS inventory_movements_no_update ON inventory_ops.movements;
DROP FUNCTION IF EXISTS inventory_ops.prevent_movement_mutation();
DROP TABLE IF EXISTS inventory_ops.movements;
DROP TABLE IF EXISTS inventory_ops.movement_numbering_series;
DROP TABLE IF EXISTS inventory_ops.locations;
DROP TABLE IF EXISTS inventory_ops.items;
DROP SCHEMA IF EXISTS inventory_ops;
