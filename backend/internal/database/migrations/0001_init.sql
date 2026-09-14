-- 0001_init: 初始 schema（与原 GORM AutoMigrate 产物一致，并修正 wafers.wafer_id 类型/外键）
-- 仅会在空库上执行；既有库通过基线机制跳过本文件。

-- ---------- 样品域 ----------

CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    username   VARCHAR(64) NOT NULL,
    full_name  VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_users_username ON users (username);
CREATE INDEX idx_users_deleted_at ON users (deleted_at);

CREATE TABLE locations (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    type        VARCHAR(16) NOT NULL,
    building    VARCHAR(64),
    room        VARCHAR(64),
    temperature NUMERIC,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_locations_name ON locations (name);
CREATE INDEX idx_locations_type ON locations (type);
CREATE INDEX idx_locations_deleted_at ON locations (deleted_at);

CREATE TABLE daily_seqs (
    day        VARCHAR(8) PRIMARY KEY,
    value      BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ
);

CREATE TABLE samples (
    id             BIGSERIAL PRIMARY KEY,
    code           VARCHAR(32) NOT NULL,
    name           VARCHAR(128) NOT NULL,
    category       VARCHAR(64),
    source         VARCHAR(128),
    receiver_id    BIGINT,
    received_at    TIMESTAMPTZ NOT NULL,
    status         VARCHAR(16) NOT NULL DEFAULT 'received',
    current_loc_id BIGINT,
    remark         TEXT,
    created_at     TIMESTAMPTZ,
    updated_at     TIMESTAMPTZ,
    deleted_at     TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_samples_code ON samples (code);
CREATE INDEX idx_samples_receiver_id ON samples (receiver_id);
CREATE INDEX idx_samples_status ON samples (status);
CREATE INDEX idx_samples_current_loc_id ON samples (current_loc_id);
CREATE INDEX idx_samples_deleted_at ON samples (deleted_at);

CREATE TABLE transfers (
    id          BIGSERIAL PRIMARY KEY,
    sample_id   BIGINT NOT NULL,
    from_loc_id BIGINT,
    to_loc_id   BIGINT NOT NULL,
    operator_id BIGINT NOT NULL,
    action      VARCHAR(32) NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    note        TEXT,
    created_at  TIMESTAMPTZ
);
CREATE INDEX idx_transfers_sample_id ON transfers (sample_id);
CREATE INDEX idx_transfers_from_loc_id ON transfers (from_loc_id);
CREATE INDEX idx_transfers_to_loc_id ON transfers (to_loc_id);
CREATE INDEX idx_transfers_operator_id ON transfers (operator_id);
CREATE INDEX idx_transfers_occurred_at ON transfers (occurred_at);

CREATE TABLE test_results (
    id           BIGSERIAL PRIMARY KEY,
    sample_id    BIGINT NOT NULL,
    item_name    VARCHAR(128) NOT NULL,
    method       VARCHAR(128),
    instrument   VARCHAR(128),
    result_value VARCHAR(256),
    unit         VARCHAR(32),
    conclusion   VARCHAR(32),
    report_no    VARCHAR(64),
    analyst_id   BIGINT,
    tested_at    TIMESTAMPTZ,
    created_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ
);
CREATE INDEX idx_test_results_sample_id ON test_results (sample_id);
CREATE INDEX idx_test_results_report_no ON test_results (report_no);
CREATE INDEX idx_test_results_analyst_id ON test_results (analyst_id);
CREATE INDEX idx_test_results_deleted_at ON test_results (deleted_at);

-- ---------- 生产域 ----------

CREATE TABLE processes (
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(32) NOT NULL,
    name       VARCHAR(64) NOT NULL,
    seq        BIGINT DEFAULT 0,
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_processes_code ON processes (code);
CREATE INDEX idx_processes_deleted_at ON processes (deleted_at);

CREATE TABLE machines (
    id             BIGSERIAL PRIMARY KEY,
    code           VARCHAR(32) NOT NULL,
    name           VARCHAR(64) NOT NULL,
    process_id     BIGINT NOT NULL,
    status         VARCHAR(16) NOT NULL DEFAULT 'idle',
    current_lot_id BIGINT,
    remark         TEXT,
    created_at     TIMESTAMPTZ,
    updated_at     TIMESTAMPTZ,
    deleted_at     TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_machines_code ON machines (code);
CREATE INDEX idx_machines_process_id ON machines (process_id);
CREATE INDEX idx_machines_status ON machines (status);
CREATE INDEX idx_machines_current_lot_id ON machines (current_lot_id);
CREATE INDEX idx_machines_deleted_at ON machines (deleted_at);

CREATE TABLE wafer_lots (
    id                  BIGSERIAL PRIMARY KEY,
    lot_no              VARCHAR(32) NOT NULL,
    product             VARCHAR(128),
    wafer_count         BIGINT NOT NULL DEFAULT 0,
    status              VARCHAR(16) NOT NULL DEFAULT 'pending',
    current_process_id  BIGINT,
    received_at         TIMESTAMPTZ NOT NULL,
    remark              TEXT,
    created_at          TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ,
    deleted_at          TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_wafer_lots_lot_no ON wafer_lots (lot_no);
CREATE INDEX idx_wafer_lots_status ON wafer_lots (status);
CREATE INDEX idx_wafer_lots_current_process_id ON wafer_lots (current_process_id);
CREATE INDEX idx_wafer_lots_deleted_at ON wafer_lots (deleted_at);

CREATE TABLE lot_daily_seqs (
    day        VARCHAR(8) PRIMARY KEY,
    value      BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ
);

CREATE TABLE lot_process_records (
    id           BIGSERIAL PRIMARY KEY,
    lot_id       BIGINT NOT NULL,
    process_id   BIGINT NOT NULL,
    machine_id   BIGINT NOT NULL,
    operator_id  BIGINT NOT NULL,
    start_time   TIMESTAMPTZ NOT NULL,
    end_time     TIMESTAMPTZ,
    duration_sec BIGINT NOT NULL DEFAULT 0,
    result       VARCHAR(16) NOT NULL DEFAULT 'pending',
    wafer_out    BIGINT NOT NULL DEFAULT 0,
    remark       TEXT,
    created_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ
);
CREATE INDEX idx_lot_process_records_lot_id ON lot_process_records (lot_id);
CREATE INDEX idx_lot_process_records_process_id ON lot_process_records (process_id);
CREATE INDEX idx_lot_process_records_machine_id ON lot_process_records (machine_id);
CREATE INDEX idx_lot_process_records_operator_id ON lot_process_records (operator_id);
CREATE INDEX idx_lot_process_records_start_time ON lot_process_records (start_time);
CREATE INDEX idx_lot_process_records_end_time ON lot_process_records (end_time);
CREATE INDEX idx_lot_process_records_result ON lot_process_records (result);

CREATE TABLE machine_status_logs (
    id          BIGSERIAL PRIMARY KEY,
    machine_id  BIGINT NOT NULL,
    from_status VARCHAR(16),
    to_status   VARCHAR(16) NOT NULL,
    operator_id BIGINT,
    reason      TEXT,
    lot_id      BIGINT,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ
);
CREATE INDEX idx_machine_status_logs_machine_id ON machine_status_logs (machine_id);
CREATE INDEX idx_machine_status_logs_operator_id ON machine_status_logs (operator_id);
CREATE INDEX idx_machine_status_logs_lot_id ON machine_status_logs (lot_id);
CREATE INDEX idx_machine_status_logs_occurred_at ON machine_status_logs (occurred_at);

CREATE TABLE wafers (
    id             BIGSERIAL PRIMARY KEY,
    lot_id         BIGINT NOT NULL,
    slot_no        BIGINT NOT NULL,
    wafer_id       VARCHAR(64),
    product        VARCHAR(128),
    current_map_id BIGINT,
    remark         TEXT,
    created_at     TIMESTAMPTZ,
    updated_at     TIMESTAMPTZ,
    deleted_at     TIMESTAMPTZ
);
CREATE INDEX idx_wafers_lot_id ON wafers (lot_id);
CREATE INDEX idx_wafers_current_map_id ON wafers (current_map_id);
CREATE INDEX idx_wafers_deleted_at ON wafers (deleted_at);

CREATE TABLE wafer_bin_maps (
    id          BIGSERIAL PRIMARY KEY,
    wafer_id    BIGINT NOT NULL,
    version     BIGINT NOT NULL DEFAULT 1,
    file_name   VARCHAR(256),
    file_path   VARCHAR(512),
    file_size   BIGINT,
    map_data    JSONB NOT NULL,
    rows        BIGINT NOT NULL,
    cols        BIGINT NOT NULL,
    notch       VARCHAR(8) NOT NULL DEFAULT 'down',
    bin_defs    JSONB NOT NULL DEFAULT '[]',
    total_dies  BIGINT NOT NULL DEFAULT 0,
    pass_dies   BIGINT NOT NULL DEFAULT 0,
    bin_summary JSONB NOT NULL DEFAULT '[]',
    yield       NUMERIC NOT NULL DEFAULT 0,
    operator_id BIGINT,
    remark      TEXT,
    created_at  TIMESTAMPTZ
);
CREATE INDEX idx_wafer_bin_maps_wafer_id ON wafer_bin_maps (wafer_id);
CREATE INDEX idx_wafer_bin_maps_operator_id ON wafer_bin_maps (operator_id);

-- ---------- 外键 ----------

ALTER TABLE samples ADD CONSTRAINT fk_samples_receiver FOREIGN KEY (receiver_id) REFERENCES users (id);
ALTER TABLE samples ADD CONSTRAINT fk_samples_current_loc FOREIGN KEY (current_loc_id) REFERENCES locations (id);
ALTER TABLE transfers ADD CONSTRAINT fk_transfers_sample FOREIGN KEY (sample_id) REFERENCES samples (id);
ALTER TABLE transfers ADD CONSTRAINT fk_transfers_from_loc FOREIGN KEY (from_loc_id) REFERENCES locations (id);
ALTER TABLE transfers ADD CONSTRAINT fk_transfers_to_loc FOREIGN KEY (to_loc_id) REFERENCES locations (id);
ALTER TABLE transfers ADD CONSTRAINT fk_transfers_operator FOREIGN KEY (operator_id) REFERENCES users (id);
ALTER TABLE test_results ADD CONSTRAINT fk_test_results_sample FOREIGN KEY (sample_id) REFERENCES samples (id);
ALTER TABLE test_results ADD CONSTRAINT fk_test_results_analyst FOREIGN KEY (analyst_id) REFERENCES users (id);

ALTER TABLE machines ADD CONSTRAINT fk_machines_process FOREIGN KEY (process_id) REFERENCES processes (id);
ALTER TABLE machines ADD CONSTRAINT fk_machines_current_lot FOREIGN KEY (current_lot_id) REFERENCES wafer_lots (id);
ALTER TABLE wafer_lots ADD CONSTRAINT fk_wafer_lots_current_process FOREIGN KEY (current_process_id) REFERENCES processes (id);
ALTER TABLE lot_process_records ADD CONSTRAINT fk_lot_process_records_lot FOREIGN KEY (lot_id) REFERENCES wafer_lots (id);
ALTER TABLE lot_process_records ADD CONSTRAINT fk_lot_process_records_process FOREIGN KEY (process_id) REFERENCES processes (id);
ALTER TABLE lot_process_records ADD CONSTRAINT fk_lot_process_records_machine FOREIGN KEY (machine_id) REFERENCES machines (id);
ALTER TABLE lot_process_records ADD CONSTRAINT fk_lot_process_records_operator FOREIGN KEY (operator_id) REFERENCES users (id);
ALTER TABLE machine_status_logs ADD CONSTRAINT fk_machine_status_logs_machine FOREIGN KEY (machine_id) REFERENCES machines (id);
ALTER TABLE machine_status_logs ADD CONSTRAINT fk_machine_status_logs_operator FOREIGN KEY (operator_id) REFERENCES users (id);
ALTER TABLE wafers ADD CONSTRAINT fk_wafers_lot FOREIGN KEY (lot_id) REFERENCES wafer_lots (id);
ALTER TABLE wafers ADD CONSTRAINT fk_wafers_current_map FOREIGN KEY (current_map_id) REFERENCES wafer_bin_maps (id);
ALTER TABLE wafer_bin_maps ADD CONSTRAINT fk_wafer_bin_maps_wafer FOREIGN KEY (wafer_id) REFERENCES wafers (id);
ALTER TABLE wafer_bin_maps ADD CONSTRAINT fk_wafer_bin_maps_operator FOREIGN KEY (operator_id) REFERENCES users (id);
