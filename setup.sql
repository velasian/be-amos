TRUNCATE TABLE users CASCADE;
INSERT INTO users (email, password, role, created_at, updated_at) 
VALUES ('ryansyahrullah62@gmail.com', '$2a$14$nDZwhhhZtF8hxI3nyXd4meMuJMysPemlhawZ7UBncAB9Q7cvoqAaW', 'superadmin', NOW(), NOW());
