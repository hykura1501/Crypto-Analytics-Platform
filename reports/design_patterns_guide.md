# Hướng Dẫn Design Patterns Toàn Diện

> Tổng hợp từ https://refactoring.guru/design-patterns

## Mục Lục
1. [Giới Thiệu](#1-giới-thiệu)
2. [Creational Patterns](#2-creational-patterns)
3. [Structural Patterns](#3-structural-patterns)
4. [Behavioral Patterns](#4-behavioral-patterns)
5. [Khi Nào Sử Dụng](#5-khi-nào-sử-dụng)
6. [So Sánh Các Patterns](#6-so-sánh-các-patterns)

---

## 1. Giới Thiệu

**Design Patterns** là các giải pháp điển hình cho các vấn đề thường gặp trong thiết kế phần mềm. Chúng không phải là code cụ thể mà là các mô hình/blueprint để giải quyết vấn đề theo cách tối ưu.

### Phân Loại

| Nhóm | Mục đích | Số lượng |
|------|----------|----------|
| **Creational** | Cơ chế tạo đối tượng | 5 |
| **Structural** | Lắp ráp đối tượng thành cấu trúc lớn | 7 |
| **Behavioral** | Thuật toán và phân chia trách nhiệm | 10 |

---

## 2. Creational Patterns

### 2.1 Factory Method

**Intent:** Cung cấp interface để tạo đối tượng trong superclass, cho phép subclass thay đổi loại đối tượng được tạo.

**Khi nào sử dụng:**
- Không biết trước chính xác loại đối tượng cần tạo
- Muốn cung cấp cách mở rộng cho thư viện/framework
- Muốn tiết kiệm tài nguyên bằng cách tái sử dụng đối tượng

**Cấu trúc:**
```
┌─────────────┐         ┌─────────────┐
│   Creator   │────────>│   Product   │
│(interface)  │         │ (interface) │
├─────────────┤         └─────────────┘
│+createProduct()│              ▲
└─────────────┘                 │
       ▲                        │
       │                        │
┌──────┴──────┐         ┌───────┴───────┐
│ConcreteCreatorA│       │ConcreteProductA│
│ConcreteCreatorB│       │ConcreteProductB│
└─────────────┘         └───────────────┘
```

**Ví dụ trong Crypto Platform:**
- `DataSourceFactory` tạo các nguồn dữ liệu khác nhau (Binance, CoinGecko, Yahoo Finance)

---

### 2.2 Abstract Factory

**Intent:** Tạo families of related objects mà không chỉ định concrete class.

**Khi nào sử dụng:**
- Code cần làm việc với nhiều family of related products
- Muốn đảm bảo các sản phẩm trong một family tương thích với nhau

**Cấu trúc:**
```
┌────────────────────┐
│  AbstractFactory   │
├────────────────────┤
│+createProductA()   │
│+createProductB()   │
└────────────────────┘
         ▲
    ┌────┴────┐
    │         │
┌───┴───┐ ┌───┴───┐
│Factory1│ │Factory2│
└───────┘ └───────┘
```

**Ví dụ trong Crypto Platform:**
- `AlertFactory` tạo bộ alerts phù hợp (Email + Slack + Push cho Premium, chỉ Email cho Free)

---

### 2.3 Builder

**Intent:** Xây dựng complex objects từng bước, cho phép tạo các biến thể khác nhau.

**Khi nào sử dụng:**
- Constructor có quá nhiều tham số
- Cần tạo các biến thể khác nhau của một sản phẩm
- Muốn tách biệt construction và representation

**Cấu trúc:**
```
Client ──> Director ──> Builder ──> Product
                          ▲
                    ┌─────┴─────┐
              ConcreteBuilder1  ConcreteBuilder2
```

**Ví dụ trong Crypto Platform:**
- `ChartBuilder` để xây dựng biểu đồ phức tạp với nhiều indicators

---

### 2.4 Prototype

**Intent:** Copy existing objects mà không phụ thuộc vào class của chúng.

**Khi nào sử dụng:**
- Muốn clone object mà không coupling với concrete class
- Giảm số lượng subclasses
- Tạo object phức tạp nhanh hơn bằng cách copy

**Ví dụ trong Crypto Platform:**
- Clone cấu hình trading strategy để tùy chỉnh

---

### 2.5 Singleton

**Intent:** Đảm bảo class chỉ có một instance và cung cấp global access point.

**Khi nào sử dụng:**
- Cần chính xác một instance (database connection, config manager)
- Cần strict control over global variables

**Cấu trúc:**
```go
type Singleton struct {
    // fields
}

var instance *Singleton
var once sync.Once

func GetInstance() *Singleton {
    once.Do(func() {
        instance = &Singleton{}
    })
    return instance
}
```

**Ví dụ trong Crypto Platform:**
- `ConfigManager` - quản lý cấu hình ứng dụng
- `WebSocketHub` - quản lý tất cả connections
- `DatabaseConnection` - pool kết nối database

---

## 3. Structural Patterns

### 3.1 Adapter

**Intent:** Cho phép objects với incompatible interfaces làm việc cùng nhau.

**Khi nào sử dụng:**
- Muốn sử dụng class có sẵn nhưng interface không tương thích
- Tạo reusable class làm việc với các class không liên quan

**Cấu trúc:**
```
┌────────┐      ┌─────────┐      ┌─────────┐
│ Client │─────>│ Adapter │─────>│ Service │
└────────┘      │(wrapper)│      │(adaptee)│
                └─────────┘      └─────────┘
```

**Ví dụ trong Crypto Platform:**
- `BinanceAdapter`, `CoinGeckoAdapter` - chuẩn hóa dữ liệu từ các API khác nhau

---

### 3.2 Decorator

**Intent:** Thêm behaviors mới cho objects bằng cách đặt chúng trong wrapper objects.

**Khi nào sử dụng:**
- Thêm responsibilities cho objects dynamically
- Không thể extend bằng inheritance (final class)
- Cần combine nhiều behaviors

**Cấu trúc:**
```
┌───────────┐
│ Component │ <─────────────────┐
│(interface)│                   │
└───────────┘                   │
      ▲                         │
      │                         │
┌─────┴─────┐         ┌─────────┴─────────┐
│ Concrete  │         │  BaseDecorator    │
│ Component │         ├───────────────────┤
└───────────┘         │-wrappee: Component│
                      └───────────────────┘
                               ▲
                      ┌────────┴────────┐
                ConcreteDecoratorA  ConcreteDecoratorB
```

**Ví dụ trong Crypto Platform:**
- Thêm các phụ phí cho giao dịch: `PeakHourFee`, `HighVolatilityFee`

---

### 3.3 Facade

**Intent:** Cung cấp simplified interface cho library, framework, hoặc complex set of classes.

**Khi nào sử dụng:**
- Cần interface đơn giản cho subsystem phức tạp
- Muốn structure subsystem thành layers

**Ví dụ trong Crypto Platform:**
- `TradingFacade` - đơn giản hóa việc gọi MarketService, AIService, AccountService

---

### 3.4 Proxy

**Intent:** Cung cấp substitute hoặc placeholder cho object khác, kiểm soát access.

**Các loại Proxy:**
- **Virtual Proxy:** Lazy initialization
- **Protection Proxy:** Access control
- **Remote Proxy:** Local execution của remote service
- **Logging Proxy:** Ghi log requests
- **Caching Proxy:** Cache kết quả

**Ví dụ trong Crypto Platform:**
- `CachingPriceProxy` - cache giá để giảm API calls
- `AuthProxy` - kiểm tra quyền trước khi gọi service

---

## 4. Behavioral Patterns

### 4.1 Observer

**Intent:** Định nghĩa subscription mechanism để notify nhiều objects về events.

**Khi nào sử dụng:**
- Thay đổi state của một object cần notify các objects khác
- Không biết trước hoặc thay đổi động set of observers

**Cấu trúc:**
```
┌───────────────┐         ┌────────────────┐
│   Publisher   │────────>│   Subscriber   │
├───────────────┤         │  (interface)   │
│+subscribe()   │         ├────────────────┤
│+unsubscribe() │         │+update()       │
│+notify()      │         └────────────────┘
└───────────────┘                  ▲
                          ┌────────┴────────┐
                    ConcreteSubscriberA  ConcreteSubscriberB
```

**Ví dụ trong Crypto Platform:**
- `PriceUpdateObserver` - notify khi giá thay đổi
- `NewsAlertObserver` - notify khi có tin tức quan trọng

---

### 4.2 Strategy

**Intent:** Định nghĩa family of algorithms, đặt mỗi cái vào separate class, làm chúng interchangeable.

**Khi nào sử dụng:**
- Muốn sử dụng different variants of an algorithm
- Có nhiều similar classes chỉ khác nhau về behavior
- Muốn isolate business logic khỏi implementation details

**Cấu trúc:**
```
┌─────────┐         ┌──────────────┐
│ Context │────────>│   Strategy   │
├─────────┤         │ (interface)  │
│+execute()│        ├──────────────┤
└─────────┘         │+execute()    │
                    └──────────────┘
                           ▲
                    ┌──────┴──────┐
              StrategyA      StrategyB
```

**Ví dụ trong Crypto Platform:**
- `PredictionStrategy` - các thuật toán dự đoán khác nhau (XGBoost, LSTM, Random Forest)
- `CrawlerStrategy` - các phương pháp crawl khác nhau (RSS, HTML Scraping, API)

---

### 4.3 Command

**Intent:** Chuyển request thành stand-alone object chứa tất cả thông tin về request.

**Khi nào sử dụng:**
- Parameterize objects với operations
- Queue operations, schedule execution, hoặc execute remotely
- Implement reversible operations (undo)

**Ví dụ trong Crypto Platform:**
- `PlaceOrderCommand`, `CancelOrderCommand` với undo functionality

---

### 4.4 Template Method

**Intent:** Định nghĩa skeleton of an algorithm trong superclass, cho subclasses override specific steps.

**Ví dụ trong Crypto Platform:**
- `BaseAnalyzer` với các bước: `loadData()` → `preprocess()` → `analyze()` → `postprocess()`

---

## 5. Khi Nào Sử Dụng

### Decision Matrix

| Vấn đề | Pattern phù hợp |
|--------|-----------------|
| Tạo object mà không biết exact class | Factory Method |
| Tạo families of related objects | Abstract Factory |
| Xây dựng object phức tạp từng bước | Builder |
| Copy object mà không coupling | Prototype |
| Chỉ cần một instance | Singleton |
| Làm việc với incompatible interfaces | Adapter |
| Thêm behavior dynamically | Decorator |
| Simplify complex subsystem | Facade |
| Control access to object | Proxy |
| Notify nhiều objects về events | Observer |
| Switch algorithms at runtime | Strategy |
| Parameterize với operations | Command |
| Define algorithm skeleton | Template Method |

---

## 6. So Sánh Các Patterns

### Factory Method vs Abstract Factory
| Aspect | Factory Method | Abstract Factory |
|--------|----------------|------------------|
| Tạo | Một product | Family of products |
| Dựa trên | Inheritance | Composition |
| Complexity | Đơn giản | Phức tạp hơn |

### Adapter vs Decorator vs Proxy
| Aspect | Adapter | Decorator | Proxy |
|--------|---------|-----------|-------|
| Interface | Thay đổi | Giữ nguyên/mở rộng | Giữ nguyên |
| Purpose | Compatibility | Add behavior | Control access |
| Wrapping | Single | Multiple layers | Single |

### Strategy vs State
| Aspect | Strategy | State |
|--------|----------|-------|
| Object aware | Không biết về nhau | Biết và có thể chuyển đổi |
| Who controls | Client chọn | Object tự chuyển |
| Purpose | Thay đổi algorithm | Thay đổi behavior theo state |

### Observer vs Mediator
| Aspect | Observer | Mediator |
|--------|----------|----------|
| Communication | Unidirectional | Bidirectional |
| Coupling | Loose | Centralized |
| Use case | Event notification | Complex interactions |

---

## Tham Khảo

- [Refactoring Guru - Design Patterns](https://refactoring.guru/design-patterns)
- [System Design Primer](https://github.com/donnemartin/system-design-primer)
- Gang of Four - Design Patterns: Elements of Reusable Object-Oriented Software
