--
-- PostgreSQL database dump
--

\restrict Arv82lxQzhkpicQl0shlQzxUqRKKURMSLjdsu0OcU2FMqLhhJAf5S1Nllvm7HDi

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

--
-- Name: assignment_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.assignment_state AS ENUM (
    'hold',
    'confirmed',
    'cancelled',
    'expired'
);


--
-- Name: auth_refresh_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.auth_refresh_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: auth_refresh_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.auth_refresh_tokens (
    id bigint DEFAULT nextval('public.auth_refresh_tokens_id_seq'::regclass) NOT NULL,
    token_hash text NOT NULL,
    issued_at timestamp without time zone DEFAULT now() NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    revoked_at timestamp without time zone,
    user_agent text,
    ip character varying(45),
    user_serial_id bigint NOT NULL
);


--
-- Name: locker_assignments_assignment_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.locker_assignments_assignment_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: locker_assignments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.locker_assignments (
    assignment_id bigint DEFAULT nextval('public.locker_assignments_assignment_id_seq'::regclass) NOT NULL,
    locker_id integer NOT NULL,
    state public.assignment_state NOT NULL,
    hold_expires_at timestamp without time zone,
    confirmed_at timestamp without time zone,
    released_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    user_serial_id bigint NOT NULL
);


--
-- Name: locker_info; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.locker_info (
    locker_id integer NOT NULL,
    owner_student_id character varying(20) DEFAULT NULL::character varying,
    location_id integer NOT NULL,
    owner_serial_id bigint
);


--
-- Name: locker_locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.locker_locations (
    location_id integer NOT NULL,
    name text NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    student_id character varying(20) NOT NULL,
    name character varying(100) NOT NULL,
    phone_number character varying(32) NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    serial_id bigint NOT NULL
);


--
-- Name: auth_refresh_tokens auth_refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auth_refresh_tokens
    ADD CONSTRAINT auth_refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: auth_refresh_tokens auth_refresh_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auth_refresh_tokens
    ADD CONSTRAINT auth_refresh_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: locker_assignments locker_assignments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locker_assignments
    ADD CONSTRAINT locker_assignments_pkey PRIMARY KEY (assignment_id);


--
-- Name: locker_info locker_info_owner_serial_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locker_info
    ADD CONSTRAINT locker_info_owner_serial_id_key UNIQUE (owner_serial_id);


--
-- Name: locker_info locker_info_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locker_info
    ADD CONSTRAINT locker_info_pkey PRIMARY KEY (locker_id);


--
-- Name: locker_locations locker_locations_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locker_locations
    ADD CONSTRAINT locker_locations_name_key UNIQUE (name);


--
-- Name: locker_locations locker_locations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locker_locations
    ADD CONSTRAINT locker_locations_pkey PRIMARY KEY (location_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (serial_id);


--
-- Name: users ux_users_ident; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT ux_users_ident UNIQUE (student_id, name, phone_number);


--
-- Name: idx_assignments_lookup; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_assignments_lookup ON public.locker_assignments USING btree (locker_id, state);


--
-- Name: idx_auth_refresh_tokens_user_serial_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_auth_refresh_tokens_user_serial_id ON public.auth_refresh_tokens USING btree (user_serial_id);


--
-- Name: ux_active_assignment_per_locker; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ux_active_assignment_per_locker ON public.locker_assignments USING btree (locker_id) WHERE (state = ANY (ARRAY['hold'::public.assignment_state, 'confirmed'::public.assignment_state]));


--
-- Name: ux_active_assignment_per_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ux_active_assignment_per_user ON public.locker_assignments USING btree (user_serial_id) WHERE (state = ANY (ARRAY['hold'::public.assignment_state, 'confirmed'::public.assignment_state]));


--
-- Name: locker_assignments fk_assignment_locker; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locker_assignments
    ADD CONSTRAINT fk_assignment_locker FOREIGN KEY (locker_id) REFERENCES public.locker_info(locker_id);


--
-- Name: locker_assignments fk_assignment_user_serial; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locker_assignments
    ADD CONSTRAINT fk_assignment_user_serial FOREIGN KEY (user_serial_id) REFERENCES public.users(serial_id) ON DELETE CASCADE;


--
-- Name: locker_info fk_locker_location; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locker_info
    ADD CONSTRAINT fk_locker_location FOREIGN KEY (location_id) REFERENCES public.locker_locations(location_id);


--
-- Name: locker_info fk_locker_owner_serial; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locker_info
    ADD CONSTRAINT fk_locker_owner_serial FOREIGN KEY (owner_serial_id) REFERENCES public.users(serial_id) ON DELETE SET NULL;


--
-- Name: auth_refresh_tokens fk_refresh_user_serial; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auth_refresh_tokens
    ADD CONSTRAINT fk_refresh_user_serial FOREIGN KEY (user_serial_id) REFERENCES public.users(serial_id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict Arv82lxQzhkpicQl0shlQzxUqRKKURMSLjdsu0OcU2FMqLhhJAf5S1Nllvm7HDi

