-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Stations table
CREATE TABLE stations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    city VARCHAR(100) NOT NULL,
    code VARCHAR(10) UNIQUE NOT NULL
);

-- Create index for fuzzy matching
CREATE INDEX idx_station_name_trgm ON stations USING gin (name gin_trgm_ops);
CREATE INDEX idx_station_city_trgm ON stations USING gin (city gin_trgm_ops);

-- Trains table
CREATE TABLE trains (
    id SERIAL PRIMARY KEY,
    number VARCHAR(20) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('Frecciarossa', 'Intercity', 'Regionale')),
    has_wifi BOOLEAN DEFAULT false,
    has_food BOOLEAN DEFAULT false,
    total_seats INTEGER NOT NULL
);

-- Schedules table
CREATE TABLE schedules (
    id SERIAL PRIMARY KEY,
    train_id INTEGER REFERENCES trains(id),
    origin_id INTEGER REFERENCES stations(id),
    destination_id INTEGER REFERENCES stations(id),
    departure_time TIME NOT NULL,
    arrival_time TIME NOT NULL,
    day_of_week INTEGER NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    price_base DECIMAL(10,2) NOT NULL,
    available_seats INTEGER NOT NULL
);

-- Create indexes for efficient queries
CREATE INDEX idx_schedules_origin ON schedules(origin_id);
CREATE INDEX idx_schedules_destination ON schedules(destination_id);
CREATE INDEX idx_schedules_day_of_week ON schedules(day_of_week);

-- Bookings table
CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    booking_ref VARCHAR(20) UNIQUE NOT NULL,
    schedule_id INTEGER REFERENCES schedules(id),
    booking_date DATE NOT NULL,
    passenger_count INTEGER NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'cancelled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bookings_ref ON bookings(booking_ref);
CREATE INDEX idx_bookings_date ON bookings(booking_date);
CREATE INDEX idx_bookings_status ON bookings(status);

-- Passengers table
CREATE TABLE passengers (
    id SERIAL PRIMARY KEY,
    booking_id INTEGER REFERENCES bookings(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    passenger_type VARCHAR(20) NOT NULL CHECK (passenger_type IN ('adult', 'senior', 'child', 'infant')),
    seat_number VARCHAR(10),
    price DECIMAL(10,2) NOT NULL
);

CREATE INDEX idx_passengers_booking ON passengers(booking_id);

-- Conversation history table
CREATE TABLE conversation_history (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(50) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    message TEXT NOT NULL,
    function_call JSONB,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_conversation_session ON conversation_history(session_id, timestamp);

-- Insert stations data
INSERT INTO stations (name, city, code) VALUES
('Milano Centrale', 'Milano', 'MI'),
('Roma Termini', 'Roma', 'RM'),
('Torino Porta Nuova', 'Torino', 'TO'),
('Firenze Santa Maria Novella', 'Firenze', 'FI'),
('Napoli Centrale', 'Napoli', 'NA'),
('Venezia Santa Lucia', 'Venezia', 'VE'),
('Bologna Centrale', 'Bologna', 'BO'),
('Verona Porta Nuova', 'Verona', 'VR');

-- Insert trains data
-- Frecciarossa trains (high-speed)
INSERT INTO trains (number, type, has_wifi, has_food, total_seats) VALUES
('FR9600', 'Frecciarossa', true, true, 500),
('FR9602', 'Frecciarossa', true, true, 500),
('FR9604', 'Frecciarossa', true, true, 500),
('FR9606', 'Frecciarossa', true, true, 500),
('FR9608', 'Frecciarossa', true, true, 500),
('FR9610', 'Frecciarossa', true, true, 500),
('FR9612', 'Frecciarossa', true, true, 500),
('FR9614', 'Frecciarossa', true, true, 500),
('FR8500', 'Frecciarossa', true, true, 480),
('FR8502', 'Frecciarossa', true, true, 480),
('FR8504', 'Frecciarossa', true, true, 480);

-- Intercity trains
INSERT INTO trains (number, type, has_wifi, has_food, total_seats) VALUES
('IC520', 'Intercity', true, false, 400),
('IC522', 'Intercity', true, false, 400),
('IC524', 'Intercity', true, false, 400),
('IC526', 'Intercity', true, false, 400),
('IC610', 'Intercity', true, false, 380),
('IC612', 'Intercity', true, false, 380);

-- Regionale trains
INSERT INTO trains (number, type, has_wifi, has_food, total_seats) VALUES
('RG2030', 'Regionale', false, false, 300),
('RG2032', 'Regionale', false, false, 300),
('RG2034', 'Regionale', false, false, 300),
('RG2036', 'Regionale', false, false, 300);

-- Insert schedules
-- Milano - Roma route (578 km, ~3 hours high-speed)
-- Frecciarossa: €0.15/km = €86.70 base
INSERT INTO schedules (train_id, origin_id, destination_id, departure_time, arrival_time, day_of_week, price_base, available_seats)
SELECT
    t.id,
    (SELECT id FROM stations WHERE code = 'MI'),
    (SELECT id FROM stations WHERE code = 'RM'),
    departure_time,
    arrival_time,
    dow,
    86.70,
    t.total_seats
FROM trains t
CROSS JOIN (VALUES
    ('06:00:00'::time, '09:00:00'::time, 1),
    ('06:00:00'::time, '09:00:00'::time, 2),
    ('06:00:00'::time, '09:00:00'::time, 3),
    ('06:00:00'::time, '09:00:00'::time, 4),
    ('06:00:00'::time, '09:00:00'::time, 5),
    ('07:30:00'::time, '10:30:00'::time, 1),
    ('07:30:00'::time, '10:30:00'::time, 2),
    ('07:30:00'::time, '10:30:00'::time, 3),
    ('07:30:00'::time, '10:30:00'::time, 4),
    ('07:30:00'::time, '10:30:00'::time, 5),
    ('09:00:00'::time, '12:00:00'::time, 1),
    ('09:00:00'::time, '12:00:00'::time, 2),
    ('09:00:00'::time, '12:00:00'::time, 3),
    ('09:00:00'::time, '12:00:00'::time, 4),
    ('09:00:00'::time, '12:00:00'::time, 5),
    ('11:30:00'::time, '14:30:00'::time, 1),
    ('11:30:00'::time, '14:30:00'::time, 2),
    ('11:30:00'::time, '14:30:00'::time, 3),
    ('11:30:00'::time, '14:30:00'::time, 4),
    ('11:30:00'::time, '14:30:00'::time, 5),
    ('14:00:00'::time, '17:00:00'::time, 1),
    ('14:00:00'::time, '17:00:00'::time, 2),
    ('14:00:00'::time, '17:00:00'::time, 3),
    ('14:00:00'::time, '17:00:00'::time, 4),
    ('14:00:00'::time, '17:00:00'::time, 5),
    ('16:30:00'::time, '19:30:00'::time, 1),
    ('16:30:00'::time, '19:30:00'::time, 2),
    ('16:30:00'::time, '19:30:00'::time, 3),
    ('16:30:00'::time, '19:30:00'::time, 4),
    ('16:30:00'::time, '19:30:00'::time, 5),
    ('18:00:00'::time, '21:00:00'::time, 1),
    ('18:00:00'::time, '21:00:00'::time, 2),
    ('18:00:00'::time, '21:00:00'::time, 3),
    ('18:00:00'::time, '21:00:00'::time, 4),
    ('18:00:00'::time, '21:00:00'::time, 5),
    ('20:00:00'::time, '23:00:00'::time, 1),
    ('20:00:00'::time, '23:00:00'::time, 2),
    ('20:00:00'::time, '23:00:00'::time, 3),
    ('20:00:00'::time, '23:00:00'::time, 4),
    ('20:00:00'::time, '23:00:00'::time, 5)
) AS times(departure_time, arrival_time, dow)
WHERE t.number LIKE 'FR96%'
LIMIT 40;

-- Roma - Milano (return)
INSERT INTO schedules (train_id, origin_id, destination_id, departure_time, arrival_time, day_of_week, price_base, available_seats)
SELECT
    t.id,
    (SELECT id FROM stations WHERE code = 'RM'),
    (SELECT id FROM stations WHERE code = 'MI'),
    departure_time,
    arrival_time,
    dow,
    86.70,
    t.total_seats
FROM trains t
CROSS JOIN (VALUES
    ('06:00:00'::time, '09:00:00'::time, 1),
    ('06:00:00'::time, '09:00:00'::time, 2),
    ('06:00:00'::time, '09:00:00'::time, 3),
    ('06:00:00'::time, '09:00:00'::time, 4),
    ('06:00:00'::time, '09:00:00'::time, 5),
    ('08:00:00'::time, '11:00:00'::time, 1),
    ('08:00:00'::time, '11:00:00'::time, 2),
    ('08:00:00'::time, '11:00:00'::time, 3),
    ('08:00:00'::time, '11:00:00'::time, 4),
    ('08:00:00'::time, '11:00:00'::time, 5),
    ('10:00:00'::time, '13:00:00'::time, 1),
    ('10:00:00'::time, '13:00:00'::time, 2),
    ('10:00:00'::time, '13:00:00'::time, 3),
    ('10:00:00'::time, '13:00:00'::time, 4),
    ('10:00:00'::time, '13:00:00'::time, 5),
    ('12:30:00'::time, '15:30:00'::time, 1),
    ('12:30:00'::time, '15:30:00'::time, 2),
    ('12:30:00'::time, '15:30:00'::time, 3),
    ('12:30:00'::time, '15:30:00'::time, 4),
    ('12:30:00'::time, '15:30:00'::time, 5),
    ('15:00:00'::time, '18:00:00'::time, 1),
    ('15:00:00'::time, '18:00:00'::time, 2),
    ('15:00:00'::time, '18:00:00'::time, 3),
    ('15:00:00'::time, '18:00:00'::time, 4),
    ('15:00:00'::time, '18:00:00'::time, 5),
    ('17:30:00'::time, '20:30:00'::time, 1),
    ('17:30:00'::time, '20:30:00'::time, 2),
    ('17:30:00'::time, '20:30:00'::time, 3),
    ('17:30:00'::time, '20:30:00'::time, 4),
    ('17:30:00'::time, '20:30:00'::time, 5),
    ('19:00:00'::time, '22:00:00'::time, 1),
    ('19:00:00'::time, '22:00:00'::time, 2),
    ('19:00:00'::time, '22:00:00'::time, 3),
    ('19:00:00'::time, '22:00:00'::time, 4),
    ('19:00:00'::time, '22:00:00'::time, 5),
    ('21:00:00'::time, '00:00:00'::time, 1),
    ('21:00:00'::time, '00:00:00'::time, 2),
    ('21:00:00'::time, '00:00:00'::time, 3),
    ('21:00:00'::time, '00:00:00'::time, 4),
    ('21:00:00'::time, '00:00:00'::time, 5)
) AS times(departure_time, arrival_time, dow)
WHERE t.number LIKE 'FR96%'
LIMIT 40;

-- Milano - Venezia route (267 km, ~2.5 hours)
-- Frecciarossa: €0.15/km = €40.05 base
INSERT INTO schedules (train_id, origin_id, destination_id, departure_time, arrival_time, day_of_week, price_base, available_seats)
SELECT
    t.id,
    (SELECT id FROM stations WHERE code = 'MI'),
    (SELECT id FROM stations WHERE code = 'VE'),
    departure_time,
    arrival_time,
    dow,
    40.05,
    t.total_seats
FROM trains t
CROSS JOIN (VALUES
    ('06:30:00'::time, '09:00:00'::time, 1),
    ('06:30:00'::time, '09:00:00'::time, 2),
    ('06:30:00'::time, '09:00:00'::time, 3),
    ('06:30:00'::time, '09:00:00'::time, 4),
    ('06:30:00'::time, '09:00:00'::time, 5),
    ('08:30:00'::time, '11:00:00'::time, 1),
    ('08:30:00'::time, '11:00:00'::time, 2),
    ('08:30:00'::time, '11:00:00'::time, 3),
    ('08:30:00'::time, '11:00:00'::time, 4),
    ('08:30:00'::time, '11:00:00'::time, 5),
    ('10:30:00'::time, '13:00:00'::time, 1),
    ('10:30:00'::time, '13:00:00'::time, 2),
    ('10:30:00'::time, '13:00:00'::time, 3),
    ('10:30:00'::time, '13:00:00'::time, 4),
    ('10:30:00'::time, '13:00:00'::time, 5),
    ('12:30:00'::time, '15:00:00'::time, 1),
    ('12:30:00'::time, '15:00:00'::time, 2),
    ('12:30:00'::time, '15:00:00'::time, 3),
    ('12:30:00'::time, '15:00:00'::time, 4),
    ('12:30:00'::time, '15:00:00'::time, 5),
    ('14:30:00'::time, '17:00:00'::time, 1),
    ('14:30:00'::time, '17:00:00'::time, 2),
    ('14:30:00'::time, '17:00:00'::time, 3),
    ('14:30:00'::time, '17:00:00'::time, 4),
    ('14:30:00'::time, '17:00:00'::time, 5),
    ('16:30:00'::time, '19:00:00'::time, 1),
    ('16:30:00'::time, '19:00:00'::time, 2),
    ('16:30:00'::time, '19:00:00'::time, 3),
    ('16:30:00'::time, '19:00:00'::time, 4),
    ('16:30:00'::time, '19:00:00'::time, 5),
    ('18:30:00'::time, '21:00:00'::time, 1),
    ('18:30:00'::time, '21:00:00'::time, 2),
    ('18:30:00'::time, '21:00:00'::time, 3),
    ('18:30:00'::time, '21:00:00'::time, 4),
    ('18:30:00'::time, '21:00:00'::time, 5)
) AS times(departure_time, arrival_time, dow)
WHERE t.number LIKE 'FR85%'
LIMIT 35;

-- Torino - Firenze route (339 km, ~4 hours)
-- Intercity: €0.10/km = €33.90 base
INSERT INTO schedules (train_id, origin_id, destination_id, departure_time, arrival_time, day_of_week, price_base, available_seats)
SELECT
    t.id,
    (SELECT id FROM stations WHERE code = 'TO'),
    (SELECT id FROM stations WHERE code = 'FI'),
    departure_time,
    arrival_time,
    dow,
    33.90,
    t.total_seats
FROM trains t
CROSS JOIN (VALUES
    ('07:00:00'::time, '11:00:00'::time, 1),
    ('07:00:00'::time, '11:00:00'::time, 2),
    ('07:00:00'::time, '11:00:00'::time, 3),
    ('07:00:00'::time, '11:00:00'::time, 4),
    ('07:00:00'::time, '11:00:00'::time, 5),
    ('09:00:00'::time, '13:00:00'::time, 1),
    ('09:00:00'::time, '13:00:00'::time, 2),
    ('09:00:00'::time, '13:00:00'::time, 3),
    ('09:00:00'::time, '13:00:00'::time, 4),
    ('09:00:00'::time, '13:00:00'::time, 5),
    ('13:00:00'::time, '17:00:00'::time, 1),
    ('13:00:00'::time, '17:00:00'::time, 2),
    ('13:00:00'::time, '17:00:00'::time, 3),
    ('13:00:00'::time, '17:00:00'::time, 4),
    ('13:00:00'::time, '17:00:00'::time, 5),
    ('15:00:00'::time, '19:00:00'::time, 1),
    ('15:00:00'::time, '19:00:00'::time, 2),
    ('15:00:00'::time, '19:00:00'::time, 3),
    ('15:00:00'::time, '19:00:00'::time, 4),
    ('15:00:00'::time, '19:00:00'::time, 5),
    ('17:00:00'::time, '21:00:00'::time, 1),
    ('17:00:00'::time, '21:00:00'::time, 2),
    ('17:00:00'::time, '21:00:00'::time, 3),
    ('17:00:00'::time, '21:00:00'::time, 4),
    ('17:00:00'::time, '21:00:00'::time, 5)
) AS times(departure_time, arrival_time, dow)
WHERE t.number LIKE 'IC5%'
LIMIT 25;

-- Roma - Napoli route (225 km, ~1.5 hours, high frequency)
-- Frecciarossa: €0.15/km = €33.75 base
INSERT INTO schedules (train_id, origin_id, destination_id, departure_time, arrival_time, day_of_week, price_base, available_seats)
SELECT
    t.id,
    (SELECT id FROM stations WHERE code = 'RM'),
    (SELECT id FROM stations WHERE code = 'NA'),
    departure_time,
    arrival_time,
    dow,
    33.75,
    t.total_seats
FROM trains t
CROSS JOIN (VALUES
    ('06:00:00'::time, '07:30:00'::time, 1), ('07:00:00'::time, '08:30:00'::time, 1),
    ('08:00:00'::time, '09:30:00'::time, 1), ('09:00:00'::time, '10:30:00'::time, 1),
    ('10:00:00'::time, '11:30:00'::time, 1), ('11:00:00'::time, '12:30:00'::time, 1),
    ('12:00:00'::time, '13:30:00'::time, 1), ('13:00:00'::time, '14:30:00'::time, 1),
    ('14:00:00'::time, '15:30:00'::time, 1), ('15:00:00'::time, '16:30:00'::time, 1),
    ('16:00:00'::time, '17:30:00'::time, 1), ('17:00:00'::time, '18:30:00'::time, 1),
    ('18:00:00'::time, '19:30:00'::time, 1), ('19:00:00'::time, '20:30:00'::time, 1),
    ('20:00:00'::time, '21:30:00'::time, 1), ('21:00:00'::time, '22:30:00'::time, 1),
    ('22:00:00'::time, '23:30:00'::time, 1),
    ('06:00:00'::time, '07:30:00'::time, 2), ('07:00:00'::time, '08:30:00'::time, 2),
    ('08:00:00'::time, '09:30:00'::time, 2), ('09:00:00'::time, '10:30:00'::time, 2),
    ('10:00:00'::time, '11:30:00'::time, 2), ('11:00:00'::time, '12:30:00'::time, 2),
    ('12:00:00'::time, '13:30:00'::time, 2), ('13:00:00'::time, '14:30:00'::time, 2),
    ('14:00:00'::time, '15:30:00'::time, 2), ('15:00:00'::time, '16:30:00'::time, 2),
    ('16:00:00'::time, '17:30:00'::time, 2), ('17:00:00'::time, '18:30:00'::time, 2),
    ('18:00:00'::time, '19:30:00'::time, 2), ('19:00:00'::time, '20:30:00'::time, 2),
    ('20:00:00'::time, '21:30:00'::time, 2), ('21:00:00'::time, '22:30:00'::time, 2),
    ('22:00:00'::time, '23:30:00'::time, 2),
    ('06:00:00'::time, '07:30:00'::time, 3), ('07:00:00'::time, '08:30:00'::time, 3),
    ('08:00:00'::time, '09:30:00'::time, 3), ('09:00:00'::time, '10:30:00'::time, 3),
    ('10:00:00'::time, '11:30:00'::time, 3), ('11:00:00'::time, '12:30:00'::time, 3),
    ('12:00:00'::time, '13:30:00'::time, 3), ('13:00:00'::time, '14:30:00'::time, 3),
    ('14:00:00'::time, '15:30:00'::time, 3), ('15:00:00'::time, '16:30:00'::time, 3),
    ('16:00:00'::time, '17:30:00'::time, 3), ('17:00:00'::time, '18:30:00'::time, 3),
    ('18:00:00'::time, '19:30:00'::time, 3), ('19:00:00'::time, '20:30:00'::time, 3),
    ('20:00:00'::time, '21:30:00'::time, 3), ('21:00:00'::time, '22:30:00'::time, 3),
    ('22:00:00'::time, '23:30:00'::time, 3),
    ('06:00:00'::time, '07:30:00'::time, 4), ('07:00:00'::time, '08:30:00'::time, 4),
    ('08:00:00'::time, '09:30:00'::time, 4), ('09:00:00'::time, '10:30:00'::time, 4),
    ('10:00:00'::time, '11:30:00'::time, 4), ('11:00:00'::time, '12:30:00'::time, 4),
    ('12:00:00'::time, '13:30:00'::time, 4), ('13:00:00'::time, '14:30:00'::time, 4),
    ('14:00:00'::time, '15:30:00'::time, 4), ('15:00:00'::time, '16:30:00'::time, 4),
    ('16:00:00'::time, '17:30:00'::time, 4), ('17:00:00'::time, '18:30:00'::time, 4),
    ('18:00:00'::time, '19:30:00'::time, 4), ('19:00:00'::time, '20:30:00'::time, 4),
    ('20:00:00'::time, '21:30:00'::time, 4), ('21:00:00'::time, '22:30:00'::time, 4),
    ('22:00:00'::time, '23:30:00'::time, 4),
    ('06:00:00'::time, '07:30:00'::time, 5), ('07:00:00'::time, '08:30:00'::time, 5),
    ('08:00:00'::time, '09:30:00'::time, 5), ('09:00:00'::time, '10:30:00'::time, 5),
    ('10:00:00'::time, '11:30:00'::time, 5), ('11:00:00'::time, '12:30:00'::time, 5),
    ('12:00:00'::time, '13:30:00'::time, 5), ('13:00:00'::time, '14:30:00'::time, 5),
    ('14:00:00'::time, '15:30:00'::time, 5), ('15:00:00'::time, '16:30:00'::time, 5),
    ('16:00:00'::time, '17:30:00'::time, 5), ('17:00:00'::time, '18:30:00'::time, 5),
    ('18:00:00'::time, '19:30:00'::time, 5), ('19:00:00'::time, '20:30:00'::time, 5),
    ('20:00:00'::time, '21:30:00'::time, 5), ('21:00:00'::time, '22:30:00'::time, 5),
    ('22:00:00'::time, '23:30:00'::time, 5)
) AS times(departure_time, arrival_time, dow)
WHERE t.type = 'Frecciarossa'
LIMIT 85;

-- Milano - Bologna route (218 km, ~1 hour)
-- Regionale: €0.06/km = €13.08 base (simplified to hourly)
INSERT INTO schedules (train_id, origin_id, destination_id, departure_time, arrival_time, day_of_week, price_base, available_seats)
SELECT
    t.id,
    (SELECT id FROM stations WHERE code = 'MI'),
    (SELECT id FROM stations WHERE code = 'BO'),
    departure_time,
    arrival_time,
    dow,
    13.08,
    t.total_seats
FROM trains t
CROSS JOIN (VALUES
    ('06:00:00'::time, '07:00:00'::time, 1), ('07:00:00'::time, '08:00:00'::time, 1),
    ('08:00:00'::time, '09:00:00'::time, 1), ('09:00:00'::time, '10:00:00'::time, 1),
    ('10:00:00'::time, '11:00:00'::time, 1), ('11:00:00'::time, '12:00:00'::time, 1),
    ('12:00:00'::time, '13:00:00'::time, 1), ('13:00:00'::time, '14:00:00'::time, 1),
    ('14:00:00'::time, '15:00:00'::time, 1), ('15:00:00'::time, '16:00:00'::time, 1),
    ('16:00:00'::time, '17:00:00'::time, 1), ('17:00:00'::time, '18:00:00'::time, 1),
    ('18:00:00'::time, '19:00:00'::time, 1), ('19:00:00'::time, '20:00:00'::time, 1),
    ('20:00:00'::time, '21:00:00'::time, 1), ('21:00:00'::time, '22:00:00'::time, 1),
    ('22:00:00'::time, '23:00:00'::time, 1), ('23:00:00'::time, '00:00:00'::time, 1),
    ('06:00:00'::time, '07:00:00'::time, 2), ('07:00:00'::time, '08:00:00'::time, 2),
    ('08:00:00'::time, '09:00:00'::time, 2), ('09:00:00'::time, '10:00:00'::time, 2),
    ('10:00:00'::time, '11:00:00'::time, 2), ('11:00:00'::time, '12:00:00'::time, 2),
    ('12:00:00'::time, '13:00:00'::time, 2), ('13:00:00'::time, '14:00:00'::time, 2),
    ('14:00:00'::time, '15:00:00'::time, 2), ('15:00:00'::time, '16:00:00'::time, 2),
    ('16:00:00'::time, '17:00:00'::time, 2), ('17:00:00'::time, '18:00:00'::time, 2),
    ('18:00:00'::time, '19:00:00'::time, 2), ('19:00:00'::time, '20:00:00'::time, 2),
    ('20:00:00'::time, '21:00:00'::time, 2), ('21:00:00'::time, '22:00:00'::time, 2),
    ('22:00:00'::time, '23:00:00'::time, 2), ('23:00:00'::time, '00:00:00'::time, 2),
    ('06:00:00'::time, '07:00:00'::time, 3), ('07:00:00'::time, '08:00:00'::time, 3),
    ('08:00:00'::time, '09:00:00'::time, 3), ('09:00:00'::time, '10:00:00'::time, 3),
    ('10:00:00'::time, '11:00:00'::time, 3), ('11:00:00'::time, '12:00:00'::time, 3),
    ('12:00:00'::time, '13:00:00'::time, 3), ('13:00:00'::time, '14:00:00'::time, 3),
    ('14:00:00'::time, '15:00:00'::time, 3), ('15:00:00'::time, '16:00:00'::time, 3),
    ('16:00:00'::time, '17:00:00'::time, 3), ('17:00:00'::time, '18:00:00'::time, 3),
    ('18:00:00'::time, '19:00:00'::time, 3), ('19:00:00'::time, '20:00:00'::time, 3),
    ('20:00:00'::time, '21:00:00'::time, 3), ('21:00:00'::time, '22:00:00'::time, 3),
    ('22:00:00'::time, '23:00:00'::time, 3), ('23:00:00'::time, '00:00:00'::time, 3),
    ('06:00:00'::time, '07:00:00'::time, 4), ('07:00:00'::time, '08:00:00'::time, 4),
    ('08:00:00'::time, '09:00:00'::time, 4), ('09:00:00'::time, '10:00:00'::time, 4),
    ('10:00:00'::time, '11:00:00'::time, 4), ('11:00:00'::time, '12:00:00'::time, 4),
    ('12:00:00'::time, '13:00:00'::time, 4), ('13:00:00'::time, '14:00:00'::time, 4),
    ('14:00:00'::time, '15:00:00'::time, 4), ('15:00:00'::time, '16:00:00'::time, 4),
    ('16:00:00'::time, '17:00:00'::time, 4), ('17:00:00'::time, '18:00:00'::time, 4),
    ('18:00:00'::time, '19:00:00'::time, 4), ('19:00:00'::time, '20:00:00'::time, 4),
    ('20:00:00'::time, '21:00:00'::time, 4), ('21:00:00'::time, '22:00:00'::time, 4),
    ('22:00:00'::time, '23:00:00'::time, 4), ('23:00:00'::time, '00:00:00'::time, 4),
    ('06:00:00'::time, '07:00:00'::time, 5), ('07:00:00'::time, '08:00:00'::time, 5),
    ('08:00:00'::time, '09:00:00'::time, 5), ('09:00:00'::time, '10:00:00'::time, 5),
    ('10:00:00'::time, '11:00:00'::time, 5), ('11:00:00'::time, '12:00:00'::time, 5),
    ('12:00:00'::time, '13:00:00'::time, 5), ('13:00:00'::time, '14:00:00'::time, 5),
    ('14:00:00'::time, '15:00:00'::time, 5), ('15:00:00'::time, '16:00:00'::time, 5),
    ('16:00:00'::time, '17:00:00'::time, 5), ('17:00:00'::time, '18:00:00'::time, 5),
    ('18:00:00'::time, '19:00:00'::time, 5), ('19:00:00'::time, '20:00:00'::time, 5),
    ('20:00:00'::time, '21:00:00'::time, 5), ('21:00:00'::time, '22:00:00'::time, 5),
    ('22:00:00'::time, '23:00:00'::time, 5), ('23:00:00'::time, '00:00:00'::time, 5)
) AS times(departure_time, arrival_time, dow)
WHERE t.number LIKE 'RG%'
LIMIT 150;

-- Verona - Venezia route (114 km, ~1.5 hours)
-- Intercity: €0.10/km = €11.40 base
INSERT INTO schedules (train_id, origin_id, destination_id, departure_time, arrival_time, day_of_week, price_base, available_seats)
SELECT
    t.id,
    (SELECT id FROM stations WHERE code = 'VR'),
    (SELECT id FROM stations WHERE code = 'VE'),
    departure_time,
    arrival_time,
    dow,
    11.40,
    t.total_seats
FROM trains t
CROSS JOIN (VALUES
    ('06:00:00'::time, '07:30:00'::time, 1), ('07:00:00'::time, '08:30:00'::time, 1),
    ('08:00:00'::time, '09:30:00'::time, 1), ('09:00:00'::time, '10:30:00'::time, 1),
    ('10:00:00'::time, '11:30:00'::time, 1), ('11:00:00'::time, '12:30:00'::time, 1),
    ('12:00:00'::time, '13:30:00'::time, 1), ('13:00:00'::time, '14:30:00'::time, 1),
    ('14:00:00'::time, '15:30:00'::time, 1), ('15:00:00'::time, '16:30:00'::time, 1),
    ('16:00:00'::time, '17:30:00'::time, 1), ('17:00:00'::time, '18:30:00'::time, 1),
    ('18:00:00'::time, '19:30:00'::time, 1), ('19:00:00'::time, '20:30:00'::time, 1),
    ('20:00:00'::time, '21:30:00'::time, 1), ('21:00:00'::time, '22:30:00'::time, 1),
    ('06:00:00'::time, '07:30:00'::time, 2), ('07:00:00'::time, '08:30:00'::time, 2),
    ('08:00:00'::time, '09:30:00'::time, 2), ('09:00:00'::time, '10:30:00'::time, 2),
    ('10:00:00'::time, '11:30:00'::time, 2), ('11:00:00'::time, '12:30:00'::time, 2),
    ('12:00:00'::time, '13:30:00'::time, 2), ('13:00:00'::time, '14:30:00'::time, 2),
    ('14:00:00'::time, '15:30:00'::time, 2), ('15:00:00'::time, '16:30:00'::time, 2),
    ('16:00:00'::time, '17:30:00'::time, 2), ('17:00:00'::time, '18:30:00'::time, 2),
    ('18:00:00'::time, '19:30:00'::time, 2), ('19:00:00'::time, '20:30:00'::time, 2),
    ('20:00:00'::time, '21:30:00'::time, 2), ('21:00:00'::time, '22:30:00'::time, 2),
    ('06:00:00'::time, '07:30:00'::time, 3), ('07:00:00'::time, '08:30:00'::time, 3),
    ('08:00:00'::time, '09:30:00'::time, 3), ('09:00:00'::time, '10:30:00'::time, 3),
    ('10:00:00'::time, '11:30:00'::time, 3), ('11:00:00'::time, '12:30:00'::time, 3),
    ('12:00:00'::time, '13:30:00'::time, 3), ('13:00:00'::time, '14:30:00'::time, 3),
    ('14:00:00'::time, '15:30:00'::time, 3), ('15:00:00'::time, '16:30:00'::time, 3),
    ('16:00:00'::time, '17:30:00'::time, 3), ('17:00:00'::time, '18:30:00'::time, 3),
    ('18:00:00'::time, '19:30:00'::time, 3), ('19:00:00'::time, '20:30:00'::time, 3),
    ('20:00:00'::time, '21:30:00'::time, 3), ('21:00:00'::time, '22:30:00'::time, 3),
    ('06:00:00'::time, '07:30:00'::time, 4), ('07:00:00'::time, '08:30:00'::time, 4),
    ('08:00:00'::time, '09:30:00'::time, 4), ('09:00:00'::time, '10:30:00'::time, 4),
    ('10:00:00'::time, '11:30:00'::time, 4), ('11:00:00'::time, '12:30:00'::time, 4),
    ('12:00:00'::time, '13:30:00'::time, 4), ('13:00:00'::time, '14:30:00'::time, 4),
    ('14:00:00'::time, '15:30:00'::time, 4), ('15:00:00'::time, '16:30:00'::time, 4),
    ('16:00:00'::time, '17:30:00'::time, 4), ('17:00:00'::time, '18:30:00'::time, 4),
    ('18:00:00'::time, '19:30:00'::time, 4), ('19:00:00'::time, '20:30:00'::time, 4),
    ('20:00:00'::time, '21:30:00'::time, 4), ('21:00:00'::time, '22:30:00'::time, 4),
    ('06:00:00'::time, '07:30:00'::time, 5), ('07:00:00'::time, '08:30:00'::time, 5),
    ('08:00:00'::time, '09:30:00'::time, 5), ('09:00:00'::time, '10:30:00'::time, 5),
    ('10:00:00'::time, '11:30:00'::time, 5), ('11:00:00'::time, '12:30:00'::time, 5),
    ('12:00:00'::time, '13:30:00'::time, 5), ('13:00:00'::time, '14:30:00'::time, 5),
    ('14:00:00'::time, '15:30:00'::time, 5), ('15:00:00'::time, '16:30:00'::time, 5),
    ('16:00:00'::time, '17:30:00'::time, 5), ('17:00:00'::time, '18:30:00'::time, 5),
    ('18:00:00'::time, '19:30:00'::time, 5), ('19:00:00'::time, '20:30:00'::time, 5),
    ('20:00:00'::time, '21:30:00'::time, 5), ('21:00:00'::time, '22:30:00'::time, 5)
) AS times(departure_time, arrival_time, dow)
WHERE t.number LIKE 'IC6%'
LIMIT 80;

-- Add weekend schedules (day_of_week 0=Sunday, 6=Saturday)
INSERT INTO schedules (train_id, origin_id, destination_id, departure_time, arrival_time, day_of_week, price_base, available_seats)
SELECT
    train_id, origin_id, destination_id, departure_time, arrival_time,
    CASE WHEN random() < 0.5 THEN 0 ELSE 6 END as day_of_week,
    price_base, available_seats
FROM schedules
WHERE day_of_week = 1
LIMIT 200;
