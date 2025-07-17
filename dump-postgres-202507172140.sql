--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5 (Homebrew)
-- Dumped by pg_dump version 17.5 (Homebrew)

-- Started on 2025-07-17 21:40:05 WIB

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

DROP DATABASE postgres;
--
-- TOC entry 3705 (class 1262 OID 5)
-- Name: postgres; Type: DATABASE; Schema: -; Owner: milhamsuryapratama
--

CREATE DATABASE postgres WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'C';


ALTER DATABASE postgres OWNER TO milhamsuryapratama;

\connect postgres

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 3706 (class 0 OID 0)
-- Dependencies: 3705
-- Name: DATABASE postgres; Type: COMMENT; Schema: -; Owner: milhamsuryapratama
--

COMMENT ON DATABASE postgres IS 'default administrative connection database';


--
-- TOC entry 4 (class 2615 OID 2200)
-- Name: public; Type: SCHEMA; Schema: -; Owner: pg_database_owner
--

CREATE SCHEMA public;


ALTER SCHEMA public OWNER TO pg_database_owner;

--
-- TOC entry 3707 (class 0 OID 0)
-- Dependencies: 4
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: pg_database_owner
--

COMMENT ON SCHEMA public IS 'standard public schema';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 219 (class 1259 OID 17281)
-- Name: users; Type: TABLE; Schema: public; Owner: milhamsuryapratama
--

CREATE TABLE public.users (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    email character varying(255)
);


ALTER TABLE public.users OWNER TO milhamsuryapratama;

--
-- TOC entry 218 (class 1259 OID 17280)
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: milhamsuryapratama
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO milhamsuryapratama;

--
-- TOC entry 3708 (class 0 OID 0)
-- Dependencies: 218
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: milhamsuryapratama
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- TOC entry 3549 (class 2604 OID 17284)
-- Name: users id; Type: DEFAULT; Schema: public; Owner: milhamsuryapratama
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- TOC entry 3699 (class 0 OID 17281)
-- Dependencies: 219
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: milhamsuryapratama
--

INSERT INTO public.users VALUES (1, 'Ilham', 'ilham@gmail.com');
INSERT INTO public.users VALUES (2, 'Hasbi', 'habis@gmail.com');
INSERT INTO public.users VALUES (3, 'Aziz', 'aziz@gmail.com');
INSERT INTO public.users VALUES (4, 'Ezyh', 'ezyh@gmail.com');
INSERT INTO public.users VALUES (5, 'Nofri', 'nofri@gmail.com');
INSERT INTO public.users VALUES (7, 'Asep', 'asep@gmail.com');
INSERT INTO public.users VALUES (8, 'Hendra', 'hendra@gmail.com');
INSERT INTO public.users VALUES (9, 'Yolo', 'yolo@gmail.com');
INSERT INTO public.users VALUES (10, 'Test', 'test@gmail.com');
INSERT INTO public.users VALUES (11, 'Test Transaction 1', 'Test Transaction 1');
INSERT INTO public.users VALUES (12, 'Test Transaction 2', 'Test Transaction 2');


--
-- TOC entry 3709 (class 0 OID 0)
-- Dependencies: 218
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: milhamsuryapratama
--

SELECT pg_catalog.setval('public.users_id_seq', 21, true);


--
-- TOC entry 3552 (class 2606 OID 17286)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: milhamsuryapratama
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- TOC entry 3550 (class 1259 OID 17296)
-- Name: users_email_idx; Type: INDEX; Schema: public; Owner: milhamsuryapratama
--

CREATE UNIQUE INDEX users_email_idx ON public.users USING btree (email);


-- Completed on 2025-07-17 21:40:05 WIB

--
-- PostgreSQL database dump complete
--

