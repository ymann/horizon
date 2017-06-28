--
-- PostgreSQL database dump
--

-- Dumped from database version 9.5.0
-- Dumped by pg_dump version 9.5.0

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

SET search_path = public, pg_catalog;

DROP INDEX IF EXISTS public.signersaccount;
DROP INDEX IF EXISTS public.sellingissuerindex;
DROP INDEX IF EXISTS public.scpenvsbyseq;
DROP INDEX IF EXISTS public.priceindex;
DROP INDEX IF EXISTS public.ledgersbyseq;
DROP INDEX IF EXISTS public.histfeebyseq;
DROP INDEX IF EXISTS public.histbyseq;
DROP INDEX IF EXISTS public.buyingissuerindex;
DROP INDEX IF EXISTS public.accountbalances;
ALTER TABLE IF EXISTS ONLY public.txhistory DROP CONSTRAINT IF EXISTS txhistory_pkey;
ALTER TABLE IF EXISTS ONLY public.txfeehistory DROP CONSTRAINT IF EXISTS txfeehistory_pkey;
ALTER TABLE IF EXISTS ONLY public.trustlines DROP CONSTRAINT IF EXISTS trustlines_pkey;
ALTER TABLE IF EXISTS ONLY public.storestate DROP CONSTRAINT IF EXISTS storestate_pkey;
ALTER TABLE IF EXISTS ONLY public.signers DROP CONSTRAINT IF EXISTS signers_pkey;
ALTER TABLE IF EXISTS ONLY public.scpquorums DROP CONSTRAINT IF EXISTS scpquorums_pkey;
ALTER TABLE IF EXISTS ONLY public.pubsub DROP CONSTRAINT IF EXISTS pubsub_pkey;
ALTER TABLE IF EXISTS ONLY public.publishqueue DROP CONSTRAINT IF EXISTS publishqueue_pkey;
ALTER TABLE IF EXISTS ONLY public.peers DROP CONSTRAINT IF EXISTS peers_pkey;
ALTER TABLE IF EXISTS ONLY public.offers DROP CONSTRAINT IF EXISTS offers_pkey;
ALTER TABLE IF EXISTS ONLY public.ledgerheaders DROP CONSTRAINT IF EXISTS ledgerheaders_pkey;
ALTER TABLE IF EXISTS ONLY public.ledgerheaders DROP CONSTRAINT IF EXISTS ledgerheaders_ledgerseq_key;
ALTER TABLE IF EXISTS ONLY public.accounts DROP CONSTRAINT IF EXISTS accounts_pkey;
ALTER TABLE IF EXISTS ONLY public.accountdata DROP CONSTRAINT IF EXISTS accountdata_pkey;
DROP TABLE IF EXISTS public.txhistory;
DROP TABLE IF EXISTS public.txfeehistory;
DROP TABLE IF EXISTS public.trustlines;
DROP TABLE IF EXISTS public.storestate;
DROP TABLE IF EXISTS public.signers;
DROP TABLE IF EXISTS public.scpquorums;
DROP TABLE IF EXISTS public.scphistory;
DROP TABLE IF EXISTS public.pubsub;
DROP TABLE IF EXISTS public.publishqueue;
DROP TABLE IF EXISTS public.peers;
DROP TABLE IF EXISTS public.offers;
DROP TABLE IF EXISTS public.ledgerheaders;
DROP TABLE IF EXISTS public.accounts;
DROP TABLE IF EXISTS public.accountdata;
DROP EXTENSION IF EXISTS plpgsql;
DROP SCHEMA IF EXISTS public;
--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA public;


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: accountdata; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE accountdata (
    accountid character varying(56) NOT NULL,
    dataname character varying(64) NOT NULL,
    datavalue character varying(112) NOT NULL
);


--
-- Name: accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE accounts (
    accountid character varying(56) NOT NULL,
    balance bigint NOT NULL,
    seqnum bigint NOT NULL,
    numsubentries integer NOT NULL,
    inflationdest character varying(56),
    homedomain character varying(32) NOT NULL,
    accounttype integer NOT NULL,
    thresholds text NOT NULL,
    flags integer NOT NULL,
    lastmodified integer NOT NULL,
    CONSTRAINT accounts_balance_check CHECK ((balance >= 0)),
    CONSTRAINT accounts_numsubentries_check CHECK ((numsubentries >= 0))
);


--
-- Name: ledgerheaders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE ledgerheaders (
    ledgerhash character(64) NOT NULL,
    prevhash character(64) NOT NULL,
    bucketlisthash character(64) NOT NULL,
    ledgerseq integer,
    closetime bigint NOT NULL,
    data text NOT NULL,
    CONSTRAINT ledgerheaders_closetime_check CHECK ((closetime >= 0)),
    CONSTRAINT ledgerheaders_ledgerseq_check CHECK ((ledgerseq >= 0))
);


--
-- Name: offers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE offers (
    sellerid character varying(56) NOT NULL,
    offerid bigint NOT NULL,
    sellingassettype integer NOT NULL,
    sellingassetcode character varying(12),
    sellingissuer character varying(56),
    buyingassettype integer NOT NULL,
    buyingassetcode character varying(12),
    buyingissuer character varying(56),
    amount bigint NOT NULL,
    pricen integer NOT NULL,
    priced integer NOT NULL,
    price double precision NOT NULL,
    flags integer NOT NULL,
    lastmodified integer NOT NULL,
    CONSTRAINT offers_amount_check CHECK ((amount >= 0)),
    CONSTRAINT offers_offerid_check CHECK ((offerid >= 0))
);


--
-- Name: peers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE peers (
    ip character varying(15) NOT NULL,
    port integer DEFAULT 0 NOT NULL,
    nextattempt timestamp without time zone NOT NULL,
    numfailures integer DEFAULT 0 NOT NULL,
    CONSTRAINT peers_numfailures_check CHECK ((numfailures >= 0)),
    CONSTRAINT peers_port_check CHECK (((port > 0) AND (port <= 65535)))
);


--
-- Name: publishqueue; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE publishqueue (
    ledger integer NOT NULL,
    state text
);


--
-- Name: pubsub; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE pubsub (
    resid character(32) NOT NULL,
    lastread integer
);


--
-- Name: scphistory; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE scphistory (
    nodeid character(56) NOT NULL,
    ledgerseq integer NOT NULL,
    envelope text NOT NULL,
    CONSTRAINT scphistory_ledgerseq_check CHECK ((ledgerseq >= 0))
);


--
-- Name: scpquorums; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE scpquorums (
    qsethash character(64) NOT NULL,
    lastledgerseq integer NOT NULL,
    qset text NOT NULL,
    CONSTRAINT scpquorums_lastledgerseq_check CHECK ((lastledgerseq >= 0))
);


--
-- Name: signers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE signers (
    accountid character varying(56) NOT NULL,
    publickey character varying(56) NOT NULL,
    weight integer NOT NULL,
    signertype integer NOT NULL
);


--
-- Name: storestate; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE storestate (
    statename character(32) NOT NULL,
    state text
);


--
-- Name: trustlines; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE trustlines (
    accountid character varying(56) NOT NULL,
    assettype integer NOT NULL,
    issuer character varying(56) NOT NULL,
    assetcode character varying(12) NOT NULL,
    tlimit bigint NOT NULL,
    balance bigint NOT NULL,
    flags integer NOT NULL,
    lastmodified integer NOT NULL,
    CONSTRAINT trustlines_balance_check CHECK ((balance >= 0)),
    CONSTRAINT trustlines_tlimit_check CHECK ((tlimit > 0))
);


--
-- Name: txfeehistory; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE txfeehistory (
    txid character(64) NOT NULL,
    ledgerseq integer NOT NULL,
    txindex integer NOT NULL,
    txchanges text NOT NULL,
    CONSTRAINT txfeehistory_ledgerseq_check CHECK ((ledgerseq >= 0))
);


--
-- Name: txhistory; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE txhistory (
    txid character(64) NOT NULL,
    ledgerseq integer NOT NULL,
    txindex integer NOT NULL,
    txbody text NOT NULL,
    txresult text NOT NULL,
    txmeta text NOT NULL,
    CONSTRAINT txhistory_ledgerseq_check CHECK ((ledgerseq >= 0))
);


--
-- Data for Name: accountdata; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: accounts; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO accounts VALUES ('GAJLXJ6AJBYG5IDQZQ45CTDYHJRZ6DI4H4IRJA6CD3W6IIJIKLPAS33R', 0, 0, 0, NULL, '', 6, 'AQAAAA==', 0, 1);
INSERT INTO accounts VALUES ('GB45Q4BIHV52PK34LB56KKU4YDELVB3NGIZECZ4257H5DDSCUXC6ADGI', 0, 0, 0, NULL, '', 6, 'AQAAAA==', 0, 1);


--
-- Data for Name: ledgerheaders; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO ledgerheaders VALUES ('0009617b7db56ad0f0fb977c6930e9f6a3bacd489d184457e8fd7ac05f6a2915', '0000000000000000000000000000000000000000000000000000000000000000', '038cea1c111272f64dc00c653bfd4d539170f375c49fc1be253687cc7d7c4c5c', 1, 0, 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADjOocERJy9k3ADGU7/U1TkXDzdcSfwb4lNofMfXxMXAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA');
INSERT INTO ledgerheaders VALUES ('8f82ea129e775ce58bedae88dbcc03f609f8a75a04b22996b282c93428a1d42f', '0009617b7db56ad0f0fb977c6930e9f6a3bacd489d184457e8fd7ac05f6a2915', '025030773993d2223cf296cbf386a45666eaf25d6ad447bfa5a55b4a480130bb', 2, 1468245508, 'AAAAAgAJYXt9tWrQ8PuXfGkw6fajus1InRhEV+j9esBfaikVgK5RGfjHJoKHu47hh6kXBFXXoVXzrveg7syj8kMeyh0AAAAAV4OmBAAAAAIAAAAIAAAAAQAAAAIAAAAIAAAAAgAAADIAAAAA3z9hmASpL9tAVxktxD3XSOp3itxSvEmM6AUkwBS4ERkCUDB3OZPSIjzylsvzhqRWZuryXWrUR7+lpVtKSAEwuwAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAyAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA');
INSERT INTO ledgerheaders VALUES ('342b37c6bc3fef074065558cfaab377682604099212e884ff1b30575965e3d28', '8f82ea129e775ce58bedae88dbcc03f609f8a75a04b22996b282c93428a1d42f', '025030773993d2223cf296cbf386a45666eaf25d6ad447bfa5a55b4a480130bb', 3, 1468245513, 'AAAAAo+C6hKed1zli+2uiNvMA/YJ+KdaBLIplrKCyTQoodQv/yLHNqaeTC0L6xeccmv/Cxmkl4yeNI2Jl4IR37knf0sAAAAAV4OmCQAAAAAAAAAA3z9hmASpL9tAVxktxD3XSOp3itxSvEmM6AUkwBS4ERkCUDB3OZPSIjzylsvzhqRWZuryXWrUR7+lpVtKSAEwuwAAAAMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAyAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA');
INSERT INTO ledgerheaders VALUES ('9d6778f2d9337c8093057f7df83d13d52d08cfe0ba33e116296f04d6e93f19d6', '342b37c6bc3fef074065558cfaab377682604099212e884ff1b30575965e3d28', 'abe62cbac7d39b662524c57ce19835f4d261e1df99a71c567048b3174325be8a', 4, 1468245518, 'AAAAAjQrN8a8P+8HQGVVjPqrN3aCYECZIS6IT/GzBXWWXj0o66bYdb+utCTk6LynRpz9y95s7OCpTuip1bE8qzZGCCYAAAAAV4OmDgAAAAAAAAAA3z9hmASpL9tAVxktxD3XSOp3itxSvEmM6AUkwBS4ERmr5iy6x9ObZiUkxXzhmDX00mHh35mnHFZwSLMXQyW+igAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAyAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA');
INSERT INTO ledgerheaders VALUES ('b618c49cbc6f32eb3127bdb1f2fe32a5404f9372cfbe69754602cbe264e95c42', '9d6778f2d9337c8093057f7df83d13d52d08cfe0ba33e116296f04d6e93f19d6', 'abe62cbac7d39b662524c57ce19835f4d261e1df99a71c567048b3174325be8a', 5, 1468245523, 'AAAAAp1nePLZM3yAkwV/ffg9E9UtCM/gujPhFilvBNbpPxnWW9jmEk580kgh1ZQ1oVIoeI3vZx8DrsmhJ9LJ8PB3DzUAAAAAV4OmEwAAAAAAAAAA3z9hmASpL9tAVxktxD3XSOp3itxSvEmM6AUkwBS4ERmr5iy6x9ObZiUkxXzhmDX00mHh35mnHFZwSLMXQyW+igAAAAUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAyAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA');


--
-- Data for Name: offers; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: peers; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: publishqueue; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: pubsub; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: scphistory; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO scphistory VALUES ('GCTI6HMWRH2QGMFKWVU5M5ZSOTKL7P7JAHZDMJJBKDHGWTEC4CJ7O3DU', 2, 'AAAAAKaPHZaJ9QMwqrVp1ncydNS/v+kB8jYlIVDOa0yC4JP3AAAAAAAAAAIAAAACAAAAAQAAAEiArlEZ+Mcmgoe7juGHqRcEVdehVfOu96DuzKPyQx7KHQAAAABXg6YEAAAAAgAAAAgAAAABAAAAAgAAAAgAAAACAAAAMgAAAAAAAAAB0Fo7rwzno9MQQb0myrnvY2OvRiYtQnENrWwmCSNbLEkAAABAQYcBpxS2IA2cWaNYh0Y83toPflEbRp8XEYoFJzlIStYXeNP5LvHs9VrePO8Z7v6jyOmROSM0zFwK4jn+0h38Aw==');
INSERT INTO scphistory VALUES ('GCTI6HMWRH2QGMFKWVU5M5ZSOTKL7P7JAHZDMJJBKDHGWTEC4CJ7O3DU', 3, 'AAAAAKaPHZaJ9QMwqrVp1ncydNS/v+kB8jYlIVDOa0yC4JP3AAAAAAAAAAMAAAACAAAAAQAAADD/Isc2pp5MLQvrF5xya/8LGaSXjJ40jYmXghHfuSd/SwAAAABXg6YJAAAAAAAAAAAAAAAB0Fo7rwzno9MQQb0myrnvY2OvRiYtQnENrWwmCSNbLEkAAABAGG6HRBLhIr+AL3kZui3fXczcaTtq46H6SWa3h26PjPxIoK8zGEuk6/qnDTGUq5F5MuowMODKMOUgEEyfPfx+DA==');
INSERT INTO scphistory VALUES ('GCTI6HMWRH2QGMFKWVU5M5ZSOTKL7P7JAHZDMJJBKDHGWTEC4CJ7O3DU', 4, 'AAAAAKaPHZaJ9QMwqrVp1ncydNS/v+kB8jYlIVDOa0yC4JP3AAAAAAAAAAQAAAACAAAAAQAAADDrpth1v660JOTovKdGnP3L3mzs4KlO6KnVsTyrNkYIJgAAAABXg6YOAAAAAAAAAAAAAAAB0Fo7rwzno9MQQb0myrnvY2OvRiYtQnENrWwmCSNbLEkAAABAQY+8ULUQ5f5wYP2ecHNeK4Gb9qUiHlfCL4qRTx5NXTVGhdKQTMgEbUc0TR7XBB5YF864K+3wIlTika6cbJNQCw==');
INSERT INTO scphistory VALUES ('GCTI6HMWRH2QGMFKWVU5M5ZSOTKL7P7JAHZDMJJBKDHGWTEC4CJ7O3DU', 5, 'AAAAAKaPHZaJ9QMwqrVp1ncydNS/v+kB8jYlIVDOa0yC4JP3AAAAAAAAAAUAAAACAAAAAQAAADBb2OYSTnzSSCHVlDWhUih4je9nHwOuyaEn0snw8HcPNQAAAABXg6YTAAAAAAAAAAAAAAAB0Fo7rwzno9MQQb0myrnvY2OvRiYtQnENrWwmCSNbLEkAAABAoVgx63hWADrTvO/4YTB8Z3wKLD34Kuvma9QR9JGhS3Mah2uYGmao4EAMKNzH5efGCAYff1/oc7Q+IinuVThnCg==');


--
-- Data for Name: scpquorums; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO scpquorums VALUES ('d05a3baf0ce7a3d31041bd26cab9ef6363af46262d42710dad6c2609235b2c49', 5, 'AAAAAQAAAAEAAAAApo8dlon1AzCqtWnWdzJ01L+/6QHyNiUhUM5rTILgk/cAAAAA');


--
-- Data for Name: signers; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: storestate; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO storestate VALUES ('databaseschema                  ', '3');
INSERT INTO storestate VALUES ('forcescponnextlaunch            ', 'false');
INSERT INTO storestate VALUES ('lastclosedledger                ', 'b618c49cbc6f32eb3127bdb1f2fe32a5404f9372cfbe69754602cbe264e95c42');
INSERT INTO storestate VALUES ('historyarchivestate             ', '{
    "version": 1,
    "server": "unknown-msvc",
    "currentLedger": 5,
    "currentBuckets": [
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "b4e97cb1ccf56173a39e69a5f55a290045ad334a3b659a66350c5b410e731be0",
            "next": {
                "state": 1,
                "output": "b4e97cb1ccf56173a39e69a5f55a290045ad334a3b659a66350c5b410e731be0"
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        {
            "curr": "0000000000000000000000000000000000000000000000000000000000000000",
            "next": {
                "state": 0
            },
            "snap": "0000000000000000000000000000000000000000000000000000000000000000"
        }
    ]
}');
INSERT INTO storestate VALUES ('lastscpdata                     ', 'AAAAAgAAAACmjx2WifUDMKq1adZ3MnTUv7/pAfI2JSFQzmtMguCT9wAAAAAAAAAFAAAAA9BaO68M56PTEEG9Jsq572Njr0YmLUJxDa1sJgkjWyxJAAAAAQAAADBb2OYSTnzSSCHVlDWhUih4je9nHwOuyaEn0snw8HcPNQAAAABXg6YTAAAAAAAAAAAAAAABAAAAMFvY5hJOfNJIIdWUNaFSKHiN72cfA67JoSfSyfDwdw81AAAAAFeDphMAAAAAAAAAAAAAAEBAt3knwzEyDYmhTu1aJAm3EH7BIkwDMJRBiKFndtQ0xs6visz/KsZTaHmCqMAmHatzhNBrKvzB5kODKE/vREsKAAAAAKaPHZaJ9QMwqrVp1ncydNS/v+kB8jYlIVDOa0yC4JP3AAAAAAAAAAUAAAACAAAAAQAAADBb2OYSTnzSSCHVlDWhUih4je9nHwOuyaEn0snw8HcPNQAAAABXg6YTAAAAAAAAAAAAAAAB0Fo7rwzno9MQQb0myrnvY2OvRiYtQnENrWwmCSNbLEkAAABAoVgx63hWADrTvO/4YTB8Z3wKLD34Kuvma9QR9JGhS3Mah2uYGmao4EAMKNzH5efGCAYff1/oc7Q+IinuVThnCgAAAAGdZ3jy2TN8gJMFf334PRPVLQjP4Loz4RYpbwTW6T8Z1gAAAAAAAAABAAAAAQAAAAEAAAAApo8dlon1AzCqtWnWdzJ01L+/6QHyNiUhUM5rTILgk/cAAAAA');


--
-- Data for Name: trustlines; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: txfeehistory; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: txhistory; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Name: accountdata_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY accountdata
    ADD CONSTRAINT accountdata_pkey PRIMARY KEY (accountid, dataname);


--
-- Name: accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (accountid);


--
-- Name: ledgerheaders_ledgerseq_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY ledgerheaders
    ADD CONSTRAINT ledgerheaders_ledgerseq_key UNIQUE (ledgerseq);


--
-- Name: ledgerheaders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY ledgerheaders
    ADD CONSTRAINT ledgerheaders_pkey PRIMARY KEY (ledgerhash);


--
-- Name: offers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY offers
    ADD CONSTRAINT offers_pkey PRIMARY KEY (offerid);


--
-- Name: peers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY peers
    ADD CONSTRAINT peers_pkey PRIMARY KEY (ip, port);


--
-- Name: publishqueue_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY publishqueue
    ADD CONSTRAINT publishqueue_pkey PRIMARY KEY (ledger);


--
-- Name: pubsub_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY pubsub
    ADD CONSTRAINT pubsub_pkey PRIMARY KEY (resid);


--
-- Name: scpquorums_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY scpquorums
    ADD CONSTRAINT scpquorums_pkey PRIMARY KEY (qsethash);


--
-- Name: signers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY signers
    ADD CONSTRAINT signers_pkey PRIMARY KEY (accountid, publickey);


--
-- Name: storestate_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY storestate
    ADD CONSTRAINT storestate_pkey PRIMARY KEY (statename);


--
-- Name: trustlines_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY trustlines
    ADD CONSTRAINT trustlines_pkey PRIMARY KEY (accountid, issuer, assetcode);


--
-- Name: txfeehistory_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY txfeehistory
    ADD CONSTRAINT txfeehistory_pkey PRIMARY KEY (ledgerseq, txindex);


--
-- Name: txhistory_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY txhistory
    ADD CONSTRAINT txhistory_pkey PRIMARY KEY (ledgerseq, txindex);


--
-- Name: accountbalances; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX accountbalances ON accounts USING btree (balance) WHERE (balance >= 1000000000);


--
-- Name: buyingissuerindex; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX buyingissuerindex ON offers USING btree (buyingissuer);


--
-- Name: histbyseq; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX histbyseq ON txhistory USING btree (ledgerseq);


--
-- Name: histfeebyseq; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX histfeebyseq ON txfeehistory USING btree (ledgerseq);


--
-- Name: ledgersbyseq; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ledgersbyseq ON ledgerheaders USING btree (ledgerseq);


--
-- Name: priceindex; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX priceindex ON offers USING btree (price);


--
-- Name: scpenvsbyseq; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX scpenvsbyseq ON scphistory USING btree (ledgerseq);


--
-- Name: sellingissuerindex; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX sellingissuerindex ON offers USING btree (sellingissuer);


--
-- Name: signersaccount; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX signersaccount ON signers USING btree (accountid);


--
-- PostgreSQL database dump complete
--

