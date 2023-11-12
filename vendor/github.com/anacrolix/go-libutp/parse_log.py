import os, sys, time

# usage: parse_log.py log-file [socket-index to focus on]


socket_filter = None
if len(sys.argv) >= 3:
    socket_filter = sys.argv[2].strip()

if socket_filter == None:
    print "scanning for socket with the most packets"
    file = open(sys.argv[1], 'rb')

    sockets = {}

    for l in file:
        if not 'our_delay' in l: continue
    
        try:
            a = l.strip().split(" ")
            socket_index = a[1][:-1]
        except:
            continue

        # msvc's runtime library doesn't prefix pointers
        # with '0x'
#        if socket_index[:2] != '0x':
#            continue

        if socket_index in sockets:
            sockets[socket_index] += 1
        else:
            sockets[socket_index] = 1

    items = sockets.items()
    items.sort(lambda x, y: y[1] - x[1])

    count = 0
    for i in items:
        print '%s: %d' % (i[0], i[1])
        count += 1
        if count > 5: break
    
    file.close()
    socket_filter = items[0][0]
    print '\nfocusing on socket %s' % socket_filter

file = open(sys.argv[1], 'rb')
out_file = 'utp.out%s' % socket_filter;
out = open(out_file, 'wb')

delay_samples = 'dots lc rgb "blue"'
delay_base = 'steps lw 2 lc rgb "purple"'
target_delay = 'steps lw 2 lc rgb "red"'
off_target = 'dots lc rgb "blue"'
cwnd = 'steps lc rgb "green"'
window_size = 'steps lc rgb "sea-green"'
rtt = 'lines lc rgb "light-blue"'

metrics = {
    'our_delay':['our delay (ms)', 'x1y2', delay_samples],
    'upload_rate':['send rate (B/s)', 'x1y1', 'lines'],
    'max_window':['cwnd (B)', 'x1y1', cwnd],
    'target_delay':['target delay (ms)', 'x1y2', target_delay],
    'cur_window':['bytes in-flight (B)', 'x1y1', window_size],
    'cur_window_packets':['number of packets in-flight', 'x1y2', 'steps'],
    'packet_size':['current packet size (B)', 'x1y2', 'steps'],
    'rtt':['rtt (ms)', 'x1y2', rtt],
    'off_target':['off-target (ms)', 'x1y2', off_target],
    'delay_sum':['delay sum (ms)', 'x1y2', 'steps'],
    'their_delay':['their delay (ms)', 'x1y2', delay_samples],
    'get_microseconds':['clock (us)', 'x1y1', 'steps'],
    'wnduser':['advertised window size (B)', 'x1y1', 'steps'],

    'delay_base':['delay base (us)', 'x1y1', delay_base],
    'their_delay_base':['their delay base (us)', 'x1y1', delay_base],
    'their_actual_delay':['their actual delay (us)', 'x1y1', delay_samples],
    'actual_delay':['actual_delay (us)', 'x1y1', delay_samples]
}

histogram_quantization = 1
socket_index = None

columns = []

begin = None

title = "-"
packet_loss = 0
packet_timeout = 0

delay_histogram = {}
window_size = {'0': 0, '1': 0}

# [35301484] 0x00ec1190: actual_delay:1021583 our_delay:102 their_delay:-1021345 off_target:297 max_window:2687 upload_rate:18942 delay_base:1021481154 delay_sum:-1021242 target_delay:400 acked_bytes:1441 cur_window:2882 scaled_gain:2.432

counter = 0

print "reading log file"

for l in file:
    if "UTP_Connect" in l:
        title = l[:-2]
        if socket_filter != None:
            title += ' socket: %s' % socket_filter
        else:
            title += ' sum of all sockets'
        continue

    try:
        a = l.strip().split(" ")
        t = a[0][1:-1]
        socket_index = a[1][:-1]
    except:
        continue
#    if socket_index[:2] != '0x':
#        continue
    
    if socket_filter != None and socket_index != socket_filter:
        continue

    counter += 1
    if (counter % 300 == 0):
        print "\r%d  " % counter,

    if "lost." in l:
        packet_loss = packet_loss + 1
        continue
    if "Packet timeout" in l:
        packet_timeout = packet_timeout + 1
        continue
    if "our_delay:" not in l:
        continue

# used for Logf timestamps
#    t, m = t.split(".")
#    t = time.strptime(t, "%H:%M:%S")
#    t = list(t)
#    t[0] += 107
#    t = tuple(t)
#    m = float(m)
#    m /= 1000.0
#    t = time.mktime(t) + m

# used for tick count timestamps
    t = int(t)

    if begin is None:
        begin = t
    t = t - begin
    # print time. Convert from milliseconds to seconds
    print >>out, '%f\t' % (float(t)/1000.),

    #if t > 200000:
    #    break

    fill_columns = not columns
    for i in a[2:]:
        try:
            n, v = i.split(':')
        except:
            continue
        v = float(v)
        if n == "our_delay":
            bucket = v / histogram_quantization
            delay_histogram[bucket] = 1 + delay_histogram.get(bucket, 0)
        if not n in metrics: continue
        if fill_columns:
            columns.append(n)
        if n == "max_window":
            window_size[socket_index] = v
            print >>out, '%f\t' % int(reduce(lambda a,b: a+b, window_size.values())),
        else:
            print >>out, '%f\t' % v,
    print >>out, float(packet_loss * 8000), float(packet_timeout * 8000)
    packet_loss = 0
    packet_timeout = 0

out.close()

out = open('%s.histogram' % out_file, 'wb')
for d,f in delay_histogram.iteritems():
    print >>out, float(d*histogram_quantization) + histogram_quantization / 2, f
out.close()


plot = [
    {
        'data': ['upload_rate', 'max_window', 'cur_window', 'wnduser', 'cur_window_packets', 'packet_size', 'rtt'],
        'title': 'send-packet-size',
        'y1': 'Bytes',
        'y2': 'Time (ms)'
    },
    {
        'data': ['our_delay', 'max_window', 'target_delay', 'cur_window', 'wnduser', 'cur_window_packets'],
        'title': 'uploading',
        'y1': 'Bytes',
        'y2': 'Time (ms)'
    },
    {
        'data': ['our_delay', 'max_window', 'target_delay', 'cur_window', 'cur_window_packets'],
        'title': 'uploading_packets',
        'y1': 'Bytes',
        'y2': 'Time (ms)'
    },
    {
        'data': ['get_microseconds'],
        'title': 'timer',
        'y1': 'Time microseconds',
        'y2': 'Time (ms)'
    },
    {
        'data': ['their_delay', 'target_delay', 'rtt'],
        'title': 'their_delay',
        'y1': '',
        'y2': 'Time (ms)'
    },
    {
        'data': ['their_actual_delay','their_delay_base'],
        'title': 'their_delay_base',
        'y1': 'Time (us)',
        'y2': ''
    },
    {
        'data': ['our_delay', 'target_delay', 'rtt'],
        'title': 'our-delay',
        'y1': '',
        'y2': 'Time (ms)'
    },
    {
        'data': ['actual_delay', 'delay_base'],
        'title': 'our_delay_base',
        'y1': 'Time (us)',
        'y2': ''
    }
]

out = open('utp.gnuplot', 'w+')

files = ''

#print >>out, 'set xtics 0, 20'
print >>out, "set term png size 1280,800"
print >>out, 'set output "%s.delays.png"' % out_file
print >>out, 'set xrange [0:250]'
print >>out, 'set xlabel "delay (ms)"'
print >>out, 'set boxwidth 1'
print >>out, 'set style fill solid'
print >>out, 'set ylabel "number of packets"'
print >>out, 'plot "%s.histogram" using 1:2 with boxes' % out_file

print >>out, "set style data steps"
#print >>out, "set yrange [0:*]"
print >>out, "set y2range [*:*]"
files += out_file + '.delays.png '
#set hidden3d
#set title "Peer bandwidth distribution"
#set xlabel "Ratio"

for p in plot:
    print >>out, 'set title "%s %s"' % (p['title'], title)
    print >>out, 'set xlabel "time (s)"'
    print >>out, 'set ylabel "%s"' % p['y1']
    print >>out, "set tics nomirror"
    print >>out, 'set y2tics'
    print >>out, 'set y2label "%s"' % p['y2']
    print >>out, 'set xrange [0:*]'
    print >>out, "set key box"
    print >>out, "set term png size 1280,800"
    print >>out, 'set output "%s-%s.png"' % (out_file, p['title'])
    files += '%s-%s.png ' % (out_file, p['title'])

    comma = ''
    print >>out, "plot",

    for c in p['data']:
        if not c in metrics: continue
        i = columns.index(c)
        print >>out, '%s"%s" using 1:%d title "%s-%s" axes %s with %s' % (comma, out_file, i + 2, metrics[c][0], metrics[c][1], metrics[c][1], metrics[c][2]),
        comma = ', '
    print >>out, ''

out.close()

os.system("gnuplot utp.gnuplot")

os.system("open %s" % files)

