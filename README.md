# gosvgreplace

this is a simple svg content replace service. It obtains some svg content from a source url, replaces hyphens and content and returns the result.

```
POST http://localhost:4211/render

{
    "source":"https://www.flaticon.com/svg/static/icons/svg/3208/3208726.svg",
    "removeHypens":true,
    "replace": {"7a8c98":"fff"}
}
```
