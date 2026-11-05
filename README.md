# BookMart Backend

Go REST API แยก package ตามรูปแบบบทเรียน: `store/`, `handler/`, `mw/`, `auth/`. ใช้ pgxpool เชื่อม PostgreSQL บน Supabase, bcrypt สำหรับรหัสผ่าน และ JWT ที่ API ออกให้

## ตั้งค่าและรัน

1. ติดตั้ง Go 1.22 ขึ้นไป และสร้าง Supabase project
2. รัน `supabase/schema.sql` แล้ว `supabase/seed.sql` ใน Supabase SQL Editor
3. ตั้ง environment variables ตาม `.env.example`: `DATABASE_URL`, `JWT_SECRET` (อย่างน้อย 32 ตัวอักษร), `CORS_ALLOWED_ORIGINS` (เริ่มต้นรวม `http://localhost:3000`)
4. จากโฟลเดอร์นี้ รัน `go run .`; API จะฟังที่ `http://localhost:8080`
5. ตั้งบัญชีสมาชิกเป็น Admin ผ่าน SQL:
   `update public.members set role='admin' where email='อีเมลทีม';`
   แล้วเข้าสู่ระบบใหม่

ห้ามใส่ secrets ใน Frontend หรือ commit ลง Git

## Deploy

สร้าง Render Web Service จากโฟลเดอร์นี้ โดยใช้ Dockerfile ที่ root; ตั้ง `DATABASE_URL`, `JWT_SECRET`, `CORS_ALLOWED_ORIGINS` (โดเมน Vercel) และ Health Check Path `/health`.

## โครงสร้าง

- `auth/`: bcrypt และ JWT
- `handler/`: HTTP routes และ request handlers
- `mw/`: Logging, Recovery, CORS, RequireLogin, RequireRole
- `store/`: SQL ผ่าน pgxpool และ transaction
- `supabase/schema.sql`, `supabase/seed.sql`: schema และข้อมูลเริ่มต้น

## ทดสอบ

รัน `go test ./...` จากโฟลเดอร์นี้
