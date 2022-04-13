CREATE TABLE public.ip_geoinfo (
    ip text NOT NULL,
    country_code text,
    country text,
    city text,
    latitude double precision,
    longitude double precision,
    mystery_value text,
    PRIMARY KEY (ip)
);

CREATE INDEX ip_geoinfo_ip_hash_index
    ON public.ip_geoinfo USING hash
    (ip)
;
