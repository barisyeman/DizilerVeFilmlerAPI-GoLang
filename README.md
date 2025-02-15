Go dilinde yazılmış MySQL veritabanı ile çalışan basit API
PUT, GET, POST işlemleri yapar JWT ile çalışır, öncelikle /login sayfasına altta belirtiğim gibi json formatında login bilgisi göndermelisiniz akabinde size jwt token döndürecek, bu token ile gerekli ekleme ve düzeltme işlemleri yapabilirsiniz.

{
  "Username": "admin",
  "Password": "password"
}

Veri ekleme formatı, ek olarak post put işlemi yaparken Head kısmını Bearer Token olarak JWT'den gelen kodu giriniz.

{
  
  "baslik": "Ezel",
  "aciklama": "açıklama.",
  "kanal": "ATV",
  "baslangic_yili": 2008,
  "bitis_yili": 2013

}
