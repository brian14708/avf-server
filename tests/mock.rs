use std::{collections::VecDeque, convert::Infallible, task::Poll};

use bytes::{BufMut, Bytes, BytesMut};
use http_body::{Body, Frame};
use tonic::{
    Request, Streaming,
    codec::{Codec, ProstCodec},
};

struct MockBody(VecDeque<Bytes>);

impl Body for MockBody {
    type Data = Bytes;
    type Error = Infallible;

    fn poll_frame(
        mut self: std::pin::Pin<&mut Self>,
        _cx: &mut std::task::Context<'_>,
    ) -> std::task::Poll<Option<Result<Frame<Self::Data>, Self::Error>>> {
        match self.0.pop_front() {
            Some(bytes) => {
                let frame = Frame::data(bytes);
                Poll::Ready(Some(Ok(frame)))
            }
            None => Poll::Ready(None),
        }
    }
}

pub fn request<T>(data: impl IntoIterator<Item = T>) -> Request<Streaming<T>>
where
    T: prost::Message + Default + 'static,
{
    let v = data
        .into_iter()
        .map(|item| {
            let mut buf = BytesMut::with_capacity(256);
            buf.put_bytes(0, 5);
            item.encode(&mut buf).unwrap();
            {
                let len = buf.len() - 5;
                let mut buf = &mut buf[1..5];
                buf.put_u32(len as u32);
            }
            buf.freeze()
        })
        .collect();
    Request::new(Streaming::new_request(
        ProstCodec::<T, T>::new().decoder(),
        MockBody(v),
        None,
        None,
    ))
}
