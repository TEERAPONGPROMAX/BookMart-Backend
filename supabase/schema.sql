-- Run once in the Supabase SQL Editor. Go connects using DATABASE_URL (pgxpool).
create table if not exists public.members (
  id bigserial primary key,
  email text not null unique,
  display_name text not null check (char_length(display_name) between 2 and 80),
  password_hash text not null,
  role text not null default 'member' check (role in ('member','admin')),
  created_at timestamptz not null default now()
);

create table if not exists public.books (
  id bigserial primary key,
  title text not null unique,
  author text not null,
  category text not null,
  price numeric(10,2) not null check (price >= 0),
  cover_url text not null default '',
  description text not null default '',
  stock integer not null default 0 check (stock >= 0),
  featured boolean not null default false,
  created_at timestamptz not null default now()
);

create table if not exists public.orders (
  id bigserial primary key,
  user_id bigint not null references public.members(id),
  status text not null default 'pending' check (status in ('pending','paid','shipped','cancelled')),
  total numeric(10,2) not null check (total >= 0),
  created_at timestamptz not null default now()
);

create table if not exists public.order_items (
  id bigserial primary key,
  order_id bigint not null references public.orders(id) on delete cascade,
  book_id bigint not null references public.books(id) on delete restrict,
  title text not null,
  quantity integer not null check (quantity > 0),
  price numeric(10,2) not null check (price >= 0)
);

create index if not exists books_category_idx on public.books(category);
create index if not exists orders_user_created_idx on public.orders(user_id, created_at desc);
create index if not exists order_items_order_idx on public.order_items(order_id);

-- Go backend uses a dedicated DB role/credential. Never put DATABASE_URL in React.
-- Promote the team's admin after registering an account:
-- update public.members set role='admin' where email='your-email@example.com';
