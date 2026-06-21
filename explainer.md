# GitWhy hoạt động thế nào (giải thích siêu dễ)

> Đọc xong cái này là hiểu toàn bộ thuật toán: lưu → tách câu → tạo vân tay → nối dây → tìm kiếm → cảnh báo.

---

## Vấn đề là gì?

AI viết code rất giỏi, nhưng nó **không nhớ** những quyết định cũ. Hôm nay nó chọn FastAPI, tuần sau nó lại tự ý thêm Express, phá vỡ thứ bạn đã thống nhất.

**GitWhy = bộ nhớ cho các quyết định.** Nó nhớ *vì sao* bạn làm gì đó, để AI sau này không phá.

---

## Hình dung: cái bảng ghim của thám tử 🧷

Tưởng tượng một tấm bảng. Bạn ghim giấy lên, rồi lấy dây nối các tờ giấy lại với nhau.

| Trong code gọi là | Thực ra là gì |
|---|---|
| `sessions` | Nguyên tờ giấy ghi chú dài (toàn bộ context). Là bằng chứng để dành. |
| `claims` | Những **câu quan trọng** xé ra từ tờ giấy đó (tối đa 7 câu). **Đây mới là ký ức thật.** |
| `edges` | Sợi **dây** nối các câu lại. Màu dây = kiểu quan hệ. |
| `claim_vectors` | 3 cái **"vân tay"** của mỗi câu, để máy đem đi so sánh. |

Tất cả cất trong 1 file: `.git/gitwhy/graph.db`.

---

## Lúc LƯU — chuyện gì xảy ra?

**Bước 1 — Agent viết tóm tắt.**
Bạn bảo agent "lưu lại đi". Agent tự viết: đã làm gì, vì sao, quyết định gì, loại bỏ gì, rủi ro gì. Rồi gọi `gitwhy_save`. *(Bước này tốn token của agent.)*

**Bước 2 — Cất nguyên tờ giấy.**
GitWhy ghi nguyên đoạn đó thành file markdown (`sessions`). Để dành làm bằng chứng. *(Không cần AI.)*

**Bước 3 — Xé ra câu quan trọng.**
GitWhy đọc và lấy ra **tối đa 7 câu cốt lõi**, ưu tiên câu quan trọng nhất:

> Quyết định chính = 5 điểm ⭐ · Phương án loại / Rủi ro = 4 · Đã làm / Lý do = 3

Ví dụ một câu: *"Dùng FastAPI cho mọi route backend, không thêm Express."* *(Chỉ là code dò từ khóa, không cần AI.)*

**Bước 4 — Dán nhãn cho mỗi câu.**
Với mỗi câu, GitWhy tự đoán thêm vài thứ (cũng bằng từ khóa):

- **scope** — câu này đụng tới đâu (thư mục `backend/**`, từ khóa `fastapi`...)
- **aliases** — tên gọi khác ("fast api", "python api server")
- **retrieval_triggers** — *tình huống tương lai* cần nhớ câu này ("đổi route backend")
- **interrupt_conditions** — *kế hoạch nguy hiểm* nên báo động ("thêm framework backend khác")

**Bước 5 — Tạo 3 "vân tay" (vector).**
Đây là chỗ hay nhất. Mỗi câu không chỉ có 1 vân tay, mà **3 cái, lấy từ 3 đoạn chữ khác nhau**:

| Vân tay | Lấy từ | Để bắt được khi bạn... |
|---|---|---|
| **claim** | chính câu đó | hỏi thẳng ("vì sao FastAPI?") |
| **retrieval** | aliases + triggers | mô tả tình huống ("tôi đang thêm route mới") |
| **interrupt** | interrupt + blast radius | nói ra kế hoạch rủi ro ("tôi sẽ thêm Express") |

Vì sao 3 cái? Vì 3 tuần sau bạn hỏi sẽ **không giống** câu lúc đầu. 3 vân tay = tìm được từ 3 góc.

**Bước 6 — Nối dây giữa các câu.**
Có 2 cách dây được nối:

*Cách 1 — cùng lần lưu, theo luật:*

| Câu A | Câu B | Dây |
|---|---|---|
| việc đã làm | một quyết định | A **IMPLEMENTS** B (thực thi) |
| ràng buộc | việc đã làm | A **CONSTRAINS** B (giới hạn) |
| quyết định | lý do | A **CAUSED_BY** B (vì) |
| quyết định | phương án bị loại | A **CONFLICTS_WITH** B (xung đột) |

*Cách 2 — khác lần lưu, theo độ giống:* câu mới đem vân tay đi so với câu cũ; giống thì nối dây mờ **RELATED_CANDIDATE** ("có thể liên quan, chưa chắc").

**Bước 7 — Cất hết vào `graph.db`.** Xong.

---

## Lúc TÌM KIẾM — đây là "node traversal" (đi theo dây)

1. Câu hỏi của bạn → biến thành 1 vân tay.
2. Đem so với tất cả vân tay trong kho, có **trọng số** (claim 1.25 > retrieval 1.0 > interrupt 0.75) → chọn câu giống nhất.
3. **Đi theo dây:** từ câu đó, lần theo các sợi dây **tối đa 2 bước** để kéo về những câu liên quan.

> Nói đơn giản: tìm thấy 1 cái ghim → đi theo dây tới các ghim xung quanh.
> Tìm ra *"Dùng FastAPI"* → kéo luôn *"...vì team chuẩn hóa Python"* (lý do) và *"...đã loại Express"* (phương án loại). Ra nguyên **chuỗi lý do**, không chỉ 1 câu.

---

## Lúc CẢNH BÁO (tripwire) — cùng máy, đổi cách chấm điểm

Trước khi agent sửa code, nó gọi `gitwhy_tripwire` kèm kế hoạch sắp làm.

1. Gộp kế hoạch thành 1 đoạn chữ → tạo vân tay.
2. So với kho, nhưng lần này **vân tay "interrupt" mạnh nhất** (1.30) → ưu tiên bắt nguy hiểm.
3. Kiểm tra thêm: cùng project không? đụng đúng thư mục không? khớp interrupt không? có dây quan hệ không?
4. Đủ dấu hiệu → bắn cảnh báo:

```
Quyết định cũ liên quan:
- Dùng FastAPI cho mọi route backend.

Vì sao quan trọng lúc này:
- Kế hoạch này đang thêm một framework backend khác.

Nên làm gì:
- Tiếp tục, sửa lại, hoặc thay thế quyết định cũ.
```

---

## Nhớ 1 câu thôi

Lưu → xé ≤7 câu → mỗi câu gắn nhãn + 3 vân tay → nối dây → **sau này** hỏi → tìm câu giống → đi 2 bước theo dây → trả về câu + cả chuỗi lý do. Tripwire y hệt, chỉ thiên về vân tay "nguy hiểm" và chạy *trước khi* sửa code.

---

## Cái nào cần AI, cái nào không?

| Bước | Chạy bằng |
|---|---|
| Agent viết tóm tắt | token của agent |
| Lưu file markdown | máy, code thường |
| Xé câu + gắn nhãn | máy, code dò từ khóa (không AI) |
| Tạo vân tay (embedding) | **chỗ duy nhất "AI"** — chạy local hoặc OpenAI |
| Tìm kiếm + cảnh báo | máy, toán vector + SQL |
| Câu cảnh báo | mẫu chữ có sẵn (không phải AI viết) |

Chỉ **embedding** là bắt buộc "AI", mà nó cũng chạy được offline. Còn lại cố tình làm bằng code thường cho rẻ và chắc.

---

*File code liên quan: [internal/graph/claims.go](internal/graph/claims.go) (xé câu + gắn nhãn + 3 vân tay), [internal/graph/graph.go](internal/graph/graph.go) (nối dây + tìm kiếm + tripwire).*
