Page({
  data: {
    imageUrls: [
      "https://img2.mukewang.com/622211a20001a95817920764.jpg",
      "https://img1.mukewang.com/62256a5d0001eb3e17920764.jpg",
      "https://img4.mukewang.com/62205ba4000134df17920764.jpg",
      "https://img2.mukewang.com/621ca25c00011f4017920764.jpg",
    ],
    current: 0,
  },

  onSwiperChanged(e: any) {
    console.log(e.detail)
  },
})
