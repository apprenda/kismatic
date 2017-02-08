if ! which python > /dev/null; then
  apt-get update -qq \
  && apt-get install -qq python2.7 \
  && ln -s /usr/bin/python2.7 /usr/bin/python
fi
