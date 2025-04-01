-- Add worker nodes to coordinator
SELECT * from master_add_node('auth-db-worker-1', 5432);
SELECT * from master_add_node('auth-db-worker-2', 5432);