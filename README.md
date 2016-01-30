# Mustache Mash

MustacheMash will be an awesome website which puts mustaches on all your Facebook friends.

# Current status

I am currently trying to implement face detection&mdash;or more specifically, mouth detection. To do this, I have created a simple template matching API that uses correlation filters to find instances of trained images inside of other images. After training this API with 17 images of various mouths, it successfully found the mouth(s) in an assortment of test images a bit less than half the time. I suspect that providing more training images, both positive and negative, will help improve this statistic.

I will continue to work on optimizing my template matching API while simultaneously experimenting with larger training sets. In the end, I hope to produce a robust face detection API that will actually be fast enough to use on a server. I may also look into combining multiple templates, since this would make the entire matching process faster.
