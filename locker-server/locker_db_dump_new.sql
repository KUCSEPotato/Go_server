--
-- PostgreSQL database dump
--

\restrict 342fwsmr3lHr9pFoW5kh4FvYmM5AqgDVt6ISBhptewS4oo3bZxt8fqzf6ePHlHk

-- Dumped from database version 16.10 (Debian 16.10-1.pgdg13+1)
-- Dumped by pg_dump version 16.10 (Debian 16.10-1.pgdg13+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

ALTER TABLE IF EXISTS ONLY public.auth_refresh_tokens DROP CONSTRAINT IF EXISTS fk_refresh_student;
ALTER TABLE IF EXISTS ONLY public.locker_info DROP CONSTRAINT IF EXISTS fk_locker_owner;
ALTER TABLE IF EXISTS ONLY public.locker_info DROP CONSTRAINT IF EXISTS fk_locker_location;
ALTER TABLE IF EXISTS ONLY public.locker_assignments DROP CONSTRAINT IF EXISTS fk_assignment_student;
ALTER TABLE IF EXISTS ONLY public.locker_assignments DROP CONSTRAINT IF EXISTS fk_assignment_locker;
DROP INDEX IF EXISTS public.ux_active_assignment_per_user;
DROP INDEX IF EXISTS public.ux_active_assignment_per_locker;
DROP INDEX IF EXISTS public.idx_assignments_lookup;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_phone_number_key;
ALTER TABLE IF EXISTS ONLY public.locker_locations DROP CONSTRAINT IF EXISTS locker_locations_pkey;
ALTER TABLE IF EXISTS ONLY public.locker_locations DROP CONSTRAINT IF EXISTS locker_locations_name_key;
ALTER TABLE IF EXISTS ONLY public.locker_info DROP CONSTRAINT IF EXISTS locker_info_pkey;
ALTER TABLE IF EXISTS ONLY public.locker_info DROP CONSTRAINT IF EXISTS locker_info_owner_key;
ALTER TABLE IF EXISTS ONLY public.locker_assignments DROP CONSTRAINT IF EXISTS locker_assignments_pkey;
ALTER TABLE IF EXISTS ONLY public.auth_refresh_tokens DROP CONSTRAINT IF EXISTS auth_refresh_tokens_token_hash_key;
ALTER TABLE IF EXISTS ONLY public.auth_refresh_tokens DROP CONSTRAINT IF EXISTS auth_refresh_tokens_pkey;
ALTER TABLE IF EXISTS public.locker_locations ALTER COLUMN location_id DROP DEFAULT;
ALTER TABLE IF EXISTS public.locker_assignments ALTER COLUMN assignment_id DROP DEFAULT;
ALTER TABLE IF EXISTS public.auth_refresh_tokens ALTER COLUMN id DROP DEFAULT;
DROP TABLE IF EXISTS public.users;
DROP SEQUENCE IF EXISTS public.locker_locations_location_id_seq;
DROP TABLE IF EXISTS public.locker_locations;
DROP TABLE IF EXISTS public.locker_info;
DROP SEQUENCE IF EXISTS public.locker_assignments_assignment_id_seq;
DROP TABLE IF EXISTS public.locker_assignments;
DROP SEQUENCE IF EXISTS public.auth_refresh_tokens_id_seq;
DROP TABLE IF EXISTS public.auth_refresh_tokens;
DROP TYPE IF EXISTS public.assignment_state;
--
-- Name: assignment_state; Type: TYPE; Schema: public; Owner: locker
--

CREATE TYPE public.assignment_state AS ENUM (
    'hold',
    'confirmed',
    'cancelled',
    'expired'
);


ALTER TYPE public.assignment_state OWNER TO locker;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: auth_refresh_tokens; Type: TABLE; Schema: public; Owner: locker
--

CREATE TABLE public.auth_refresh_tokens (
    id bigint NOT NULL,
    student_id character varying(20) NOT NULL,
    token_hash text NOT NULL,
    issued_at timestamp without time zone DEFAULT now() NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    revoked_at timestamp without time zone,
    user_agent text,
    ip character varying(45)
);


ALTER TABLE public.auth_refresh_tokens OWNER TO locker;

--
-- Name: auth_refresh_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: locker
--

CREATE SEQUENCE public.auth_refresh_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.auth_refresh_tokens_id_seq OWNER TO locker;

--
-- Name: auth_refresh_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: locker
--

ALTER SEQUENCE public.auth_refresh_tokens_id_seq OWNED BY public.auth_refresh_tokens.id;


--
-- Name: locker_assignments; Type: TABLE; Schema: public; Owner: locker
--

CREATE TABLE public.locker_assignments (
    assignment_id bigint NOT NULL,
    locker_id integer NOT NULL,
    student_id character varying(20) NOT NULL,
    state public.assignment_state NOT NULL,
    hold_expires_at timestamp without time zone,
    confirmed_at timestamp without time zone,
    released_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.locker_assignments OWNER TO locker;

--
-- Name: locker_assignments_assignment_id_seq; Type: SEQUENCE; Schema: public; Owner: locker
--

CREATE SEQUENCE public.locker_assignments_assignment_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.locker_assignments_assignment_id_seq OWNER TO locker;

--
-- Name: locker_assignments_assignment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: locker
--

ALTER SEQUENCE public.locker_assignments_assignment_id_seq OWNED BY public.locker_assignments.assignment_id;


--
-- Name: locker_info; Type: TABLE; Schema: public; Owner: locker
--

CREATE TABLE public.locker_info (
    locker_id integer NOT NULL,
    owner character varying(20) DEFAULT NULL::character varying,
    location_id integer NOT NULL
);


ALTER TABLE public.locker_info OWNER TO locker;

--
-- Name: locker_locations; Type: TABLE; Schema: public; Owner: locker
--

CREATE TABLE public.locker_locations (
    location_id integer NOT NULL,
    name text NOT NULL
);


ALTER TABLE public.locker_locations OWNER TO locker;

--
-- Name: locker_locations_location_id_seq; Type: SEQUENCE; Schema: public; Owner: locker
--

CREATE SEQUENCE public.locker_locations_location_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.locker_locations_location_id_seq OWNER TO locker;

--
-- Name: locker_locations_location_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: locker
--

ALTER SEQUENCE public.locker_locations_location_id_seq OWNED BY public.locker_locations.location_id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: locker
--

CREATE TABLE public.users (
    student_id character varying(20) NOT NULL,
    name character varying(100) NOT NULL,
    phone_number character varying(32) NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.users OWNER TO locker;

--
-- Name: auth_refresh_tokens id; Type: DEFAULT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.auth_refresh_tokens ALTER COLUMN id SET DEFAULT nextval('public.auth_refresh_tokens_id_seq'::regclass);


--
-- Name: locker_assignments assignment_id; Type: DEFAULT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.locker_assignments ALTER COLUMN assignment_id SET DEFAULT nextval('public.locker_assignments_assignment_id_seq'::regclass);


--
-- Name: locker_locations location_id; Type: DEFAULT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.locker_locations ALTER COLUMN location_id SET DEFAULT nextval('public.locker_locations_location_id_seq'::regclass);


--
-- Data for Name: auth_refresh_tokens; Type: TABLE DATA; Schema: public; Owner: locker
--

COPY public.auth_refresh_tokens (id, student_id, token_hash, issued_at, expires_at, revoked_at, user_agent, ip) FROM stdin;
1	2023320060	zBRdca9PiIm1fHmjALT9UB1BYmmXQ8zMPRWICr6cFyk	2025-09-08 07:15:44.002833	2025-09-22 07:15:44.002623	\N	Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36	192.168.65.1
\.


--
-- Data for Name: locker_assignments; Type: TABLE DATA; Schema: public; Owner: locker
--

COPY public.locker_assignments (assignment_id, locker_id, student_id, state, hold_expires_at, confirmed_at, released_at, created_at) FROM stdin;
\.


--
-- Data for Name: locker_info; Type: TABLE DATA; Schema: public; Owner: locker
--

COPY public.locker_info (locker_id, owner, location_id) FROM stdin;
101	\N	1
102	\N	1
103	\N	1
104	\N	2
105	\N	2
106	\N	2
201	\N	3
202	\N	3
301	\N	4
302	\N	4
303	\N	4
304	\N	4
401	\N	5
402	\N	5
403	\N	5
404	\N	5
501	\N	6
502	\N	6
503	\N	6
601	\N	7
602	\N	7
603	\N	7
\.


--
-- Data for Name: locker_locations; Type: TABLE DATA; Schema: public; Owner: locker
--

COPY public.locker_locations (location_id, name) FROM stdin;
1	정보관 B1 엘리베이터 옆1
2	정보관 B1 엘리베이터 옆2
3	정보관 B1 기계실 옆
4	정보관 2층
5	정보관 3층
6	과학도서관 6층 620호 옆
7	과학도서관 6층 614호 옆
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: locker
--

COPY public.users (student_id, name, phone_number, created_at, updated_at) FROM stdin;
2023321234	홍길동	01012345678	2025-09-08 07:05:47.345692	2025-09-08 07:05:47.345692
2023325678	김철수	01087654321	2025-09-08 07:05:51.967131	2025-09-08 07:05:51.967131
2023320060	이정민	01023493307	2025-09-08 07:15:19.285172	2025-09-08 07:15:19.285172
\.


--
-- Name: auth_refresh_tokens_id_seq; Type: SEQUENCE SET; Schema: public; Owner: locker
--

SELECT pg_catalog.setval('public.auth_refresh_tokens_id_seq', 1, true);


--
-- Name: locker_assignments_assignment_id_seq; Type: SEQUENCE SET; Schema: public; Owner: locker
--

SELECT pg_catalog.setval('public.locker_assignments_assignment_id_seq', 1, false);


--
-- Name: locker_locations_location_id_seq; Type: SEQUENCE SET; Schema: public; Owner: locker
--

SELECT pg_catalog.setval('public.locker_locations_location_id_seq', 7, true);


--
-- Name: auth_refresh_tokens auth_refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.auth_refresh_tokens
    ADD CONSTRAINT auth_refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: auth_refresh_tokens auth_refresh_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.auth_refresh_tokens
    ADD CONSTRAINT auth_refresh_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: locker_assignments locker_assignments_pkey; Type: CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.locker_assignments
    ADD CONSTRAINT locker_assignments_pkey PRIMARY KEY (assignment_id);


--
-- Name: locker_info locker_info_owner_key; Type: CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.locker_info
    ADD CONSTRAINT locker_info_owner_key UNIQUE (owner);


--
-- Name: locker_info locker_info_pkey; Type: CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.locker_info
    ADD CONSTRAINT locker_info_pkey PRIMARY KEY (locker_id);


--
-- Name: locker_locations locker_locations_name_key; Type: CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.locker_locations
    ADD CONSTRAINT locker_locations_name_key UNIQUE (name);


--
-- Name: locker_locations locker_locations_pkey; Type: CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.locker_locations
    ADD CONSTRAINT locker_locations_pkey PRIMARY KEY (location_id);


--
-- Name: users users_phone_number_key; Type: CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_phone_number_key UNIQUE (phone_number);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (student_id);


--
-- Name: idx_assignments_lookup; Type: INDEX; Schema: public; Owner: locker
--

CREATE INDEX idx_assignments_lookup ON public.locker_assignments USING btree (locker_id, state);


--
-- Name: ux_active_assignment_per_locker; Type: INDEX; Schema: public; Owner: locker
--

CREATE UNIQUE INDEX ux_active_assignment_per_locker ON public.locker_assignments USING btree (locker_id) WHERE (state = ANY (ARRAY['hold'::public.assignment_state, 'confirmed'::public.assignment_state]));


--
-- Name: ux_active_assignment_per_user; Type: INDEX; Schema: public; Owner: locker
--

CREATE UNIQUE INDEX ux_active_assignment_per_user ON public.locker_assignments USING btree (student_id) WHERE (state = ANY (ARRAY['hold'::public.assignment_state, 'confirmed'::public.assignment_state]));


--
-- Name: locker_assignments fk_assignment_locker; Type: FK CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.locker_assignments
    ADD CONSTRAINT fk_assignment_locker FOREIGN KEY (locker_id) REFERENCES public.locker_info(locker_id);


--
-- Name: locker_assignments fk_assignment_student; Type: FK CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.locker_assignments
    ADD CONSTRAINT fk_assignment_student FOREIGN KEY (student_id) REFERENCES public.users(student_id);


--
-- Name: locker_info fk_locker_location; Type: FK CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.locker_info
    ADD CONSTRAINT fk_locker_location FOREIGN KEY (location_id) REFERENCES public.locker_locations(location_id);


--
-- Name: locker_info fk_locker_owner; Type: FK CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.locker_info
    ADD CONSTRAINT fk_locker_owner FOREIGN KEY (owner) REFERENCES public.users(student_id);


--
-- Name: auth_refresh_tokens fk_refresh_student; Type: FK CONSTRAINT; Schema: public; Owner: locker
--

ALTER TABLE ONLY public.auth_refresh_tokens
    ADD CONSTRAINT fk_refresh_student FOREIGN KEY (student_id) REFERENCES public.users(student_id);


--
-- PostgreSQL database dump complete
--

\unrestrict 342fwsmr3lHr9pFoW5kh4FvYmM5AqgDVt6ISBhptewS4oo3bZxt8fqzf6ePHlHk

