-- 0002_fix_wafers_wafer_id: 修复 AutoMigrate 时代的 wafers.wafer_id 缺陷。
--
-- 背景：GORM 关系推断曾把 Wafer.WaferID（晶圆刻号，字符串）误当作指向
-- wafer_bin_maps 的外键，导致老库中 wafers.wafer_id 为 bigint 且带有错误外键，
-- 而 wafer_bin_maps.wafer_id 缺少指向 wafers 的外键；空字符串刻号插入 bigint
-- 列会直接报错，晶圆槽位创建 silently 失败。
--
-- 本迁移对老库（基线跳过 0001）进行矫正；对新库（0001 已是正确结构）为无害空操作。

ALTER TABLE wafers DROP CONSTRAINT IF EXISTS fk_wafer_bin_maps_wafer;

ALTER TABLE wafers ALTER COLUMN wafer_id TYPE VARCHAR(64) USING wafer_id::text;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_wafer_bin_maps_wafer' AND conrelid = 'wafer_bin_maps'::regclass
    ) THEN
        ALTER TABLE wafer_bin_maps
            ADD CONSTRAINT fk_wafer_bin_maps_wafer FOREIGN KEY (wafer_id) REFERENCES wafers (id);
    END IF;
END $$;
