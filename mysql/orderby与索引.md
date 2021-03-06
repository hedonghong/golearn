order by的字段不在where条件不在select中     有排序操作

order by的字段不在where条件但在select中     有排序操作

order by的字段在where条件但不在select中     无排序操作

order by的字段在where条件但不在select中(倒序)     无排序操作



结论：

当order by 字段出现在where条件中时，才会利用索引而无需排序操作。其他情况，order by不会出现排序操作。

分析：

为什么只有order by 字段出现在where条件中时,才会利用该字段的索引而避免排序。这要说到数据库如何取到我们需要的数据了。

一条SQL实际上可以分为三步。

1.得到数据

2.处理数据

3.返回处理后的数据

比如上面的这条语句select sid from zhuyuehua.student where sid < 50000 and id < 50000 order by id desc

第一步：根据where条件和统计信息生成执行计划，得到数据。

第二步：将得到的数据排序。

当执行处理数据（order by）时，数据库会先查看第一步的执行计划，看order by 的字段是否在执行计划中利用了索引。如果是，则可以利用索引顺序而直接取得已经排好序的数据。如果不是，则排序操作。

第三步：返回排序后的数据。

另外：

上面的5万的数据sort只用了25ms，也许大家觉得sort不怎么占用资源。可是，由于上面的表的数据是有序的，所以排序花费的时间较少。如果 是个比较无序的表，sort时间就会增加很多了。另外排序操作一般都是在内存里进行的，对于数据库来说是一种CPU的消耗，由于现在CPU的性能增强，对 于普通的几十条或上百条记录排序对系统的影响也不会很大。但是当你的记录集增加到上百万条以上时，你需要注意是否一定要这么做了，大记录集排序不仅增加了 CPU开销，而且可能会由于内存不足发生硬盘排序的现象，当发生硬盘排序时性能会急剧下降。

注：ORACLE或者DB2都有一个空间来供SORT操作使用（上面所说的内存排序），如ORACLE中是用户全局区（UGA），里面有SORT_AREA_SIZE等参数的设置。如果当排序的数据量大时，就会出现排序溢出（硬盘排序），这时的性能就会降低很多了。

总结：

当order by 中的字段出现在where条件中时，才会利用索引而不排序，更准确的说，order by 中的字段在执行计划中利用了索引时，不用排序操作。

这个结论不仅对order by有效，对其他需要排序的操作也有效。比如group by 、union 、distinct等。