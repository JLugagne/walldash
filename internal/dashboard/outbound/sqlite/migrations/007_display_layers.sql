ALTER TABLE levels ADD COLUMN layers_json TEXT NOT NULL DEFAULT '["controls","sensors"]';
ALTER TABLE device_placements ADD COLUMN layer TEXT NOT NULL DEFAULT 'controls';
