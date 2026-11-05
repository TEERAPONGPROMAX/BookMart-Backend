insert into public.books(title,author,category,price,cover_url,description,stock,featured) values
('ร้านไม่สะดวกซื้อของคุณทกโก','คิม โฮยอน','วรรณกรรม',325,'https://images.unsplash.com/photo-1544947950-fa07a98d237f?w=600&q=85','เรื่องราวอบอุ่นหัวใจของร้านสะดวกซื้อเล็ก ๆ ที่เปลี่ยนชีวิตผู้คนในย่าน',18,true),
('Atomic Habits เพราะชีวิตดีได้กว่าที่เป็น','James Clear','พัฒนาตนเอง',395,'https://images.unsplash.com/photo-1543002588-bfa74002ed7e?w=600&q=85','เปลี่ยนแปลงเล็ก ๆ ทุกวัน เพื่อผลลัพธ์ที่ยิ่งใหญ่ในระยะยาว',25,true),
('ก่อนที่กาแฟจะเย็นลง','โทชิคาซึ คาวากุจิ','วรรณกรรม',289,'https://images.unsplash.com/photo-1512820790803-83ca734da794?w=600&q=85','คาเฟ่เล็ก ๆ ที่ให้คุณย้อนเวลากลับไปพบคนสำคัญอีกครั้ง',9,false),
('จิตวิทยาสายดาร์ก','Dr. Hiro','จิตวิทยา',275,'https://images.unsplash.com/photo-1519682337058-a94d519337bc?w=600&q=85','เข้าใจเบื้องหลังความคิดและพฤติกรรมของมนุษย์',12,false),
('The Psychology of Money','Morgan Housel','ธุรกิจ',420,'https://images.unsplash.com/photo-1497633762265-9d179a990aa6?w=600&q=85','บทเรียนอมตะเกี่ยวกับความมั่งคั่ง ความโลภ และความสุข',15,false),
('เจ้าชายน้อย','Antoine de Saint-Exupéry','แฟนตาซี',195,'https://images.unsplash.com/photo-1516979187457-637abb4f9353?w=600&q=85','นิทานแสนลึกซึ้งที่จะทำให้คุณมองโลกด้วยหัวใจอีกครั้ง',30,false),
('ชีวิตเรามีแค่สี่พันสัปดาห์','Oliver Burkeman','พัฒนาตนเอง',350,'https://images.unsplash.com/photo-1517842645767-c639042777db?w=600&q=85','คู่มือจัดการเวลา สำหรับชีวิตที่ไม่อาจทำทุกอย่างได้',7,false),
('The Midnight Library','Matt Haig','แฟนตาซี',390,'https://images.unsplash.com/photo-1519682337058-a94d519337bc?w=600&q=85','ห้องสมุดระหว่างความเป็นและความตาย กับชีวิตนับไม่ถ้วน',11,false)
on conflict (title) do nothing;
