# API design 

**"TRY"** to follow Google api design <https://cloud.google.com/apis/design>


Naming should be simple, intuitive and consistent

## gRPC methods

You **are not** late. (indicative)
**Don't be** late! (imperative)

Methods names
- The verb portion of the method name should use the imperative mood, which is for 
  orders or commands rather than the indicative mood which is for questions.
- the noun portion of the method name must be singular for all methods except List.
- custome methods may be singular or plural as appropriate.
- Batch methods **must** use the plural noun.

| Verb   | Noun | Method Name      | Request Message         | Response Message          |
|--------|------|------------------|-------------------------|---------------------------|
| List   | Book | ListBooks        | ListBooksRequest        | ListBooksResponse         |
| Get    | Book | GetBook          | GetBookRequest          | Book                      |
| Create | Book | CreateBook       | CreateBookRequest       | Book                      |
| Update | Book | UpdateBook       | UpdateBookRequest       | Book                      |
| Rename | Book | RenameBook       | RenameBookRequest       | RenameBookResponse        |
| Delete | Book | DeleteBook       | DeleteBookRequest       | google.protobuf.Empty     |







> this is blockquotes
> also blockquotes

