CREATE TABLE public.products (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY, 
	name TEXT NOT NULL CHECK (btrim(name) <> ''),
	price NUMERIC NOT NULL CHECK (price >= 0),
	stock INTEGER NOT NULL CHECK (stock >= 0)
);

